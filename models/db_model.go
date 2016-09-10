package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var SqlErrNoRows = sql.ErrNoRows

type DbModel struct {
	db *sql.DB
}

func NewDbModel(db *sql.DB) *DbModel {
	return &DbModel{db: db}
}

func (model *DbModel) Reset() {
	model.deleteFrom("devices")
	model.restartSequence("devices_id_seq")
	model.deleteFrom("todo_items")
	model.restartSequence("todo_items_id_seq")
}

func (model *DbModel) deleteFrom(tableName string) {
	sql := fmt.Sprintf("DELETE FROM \"%s\"", tableName)
	_, err := model.db.Exec(sql)
	if err != nil {
		panic(fmt.Sprintf("Error from db.Exec with sql=%s: %s", sql, err))
	}
}

func (model *DbModel) restartSequence(sequenceName string) {
	sql := fmt.Sprintf("ALTER SEQUENCE \"%s\" RESTART WITH 1;", sequenceName)
	_, err := model.db.Exec(sql)
	if err != nil {
		panic(fmt.Sprintf("Error from db.Exec with sql=%s: %s", sql, err))
	}
}

func (model *DbModel) FindOrCreateDeviceByUid(uid string) Device {
	device := model.findDeviceByUid(uid)
	if device.Id != 0 {
		return device
	} else {
		model.createDeviceIgnoringDuplicate(uid)
		return model.findDeviceByUid(uid)
	}
}

func (model *DbModel) createDeviceIgnoringDuplicate(uid string) {
	sql := `INSERT INTO devices(
			uid,
			action_to_sync_id_to_output_json,
			completed_action_to_sync_id
		) VALUES(
			$1,
			'{}',
			0
		);`
	_, err := model.db.Exec(sql, uid)
	if err != nil {
		if strings.HasPrefix(err.Error(),
			"pq: duplicate key value violates unique constraint") {
			// ignore it
		} else {
			panic(fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err))
		}
	}
}

func (model *DbModel) findDeviceByUid(uid string) Device {
	var device Device
	var actionToSyncIdToOutputJson string
	sql := `SELECT id, uid, action_to_sync_id_to_output_json
		FROM devices
		WHERE uid = $1`
	err := model.db.QueryRow(sql, uid).Scan(&device.Id, &device.Uid,
		&actionToSyncIdToOutputJson)
	if err == nil {
		var actionToSyncIdToOutput map[string]int
		if err := json.Unmarshal([]byte(actionToSyncIdToOutputJson),
			&actionToSyncIdToOutput); err != nil {
			panic(fmt.Errorf("Error from unmarshaling JSON '%s': %s",
				actionToSyncIdToOutputJson, err))
		}
		device.ActionToSyncIdToOutput = mapStringIntToMapIntInt(actionToSyncIdToOutput)
		return device
	} else if err == SqlErrNoRows {
		return Device{}
	} else {
		panic(fmt.Errorf("Error from db.QueryRow with sql=%s: %s", sql, err))
	}
}

func (model *DbModel) CreateTodo(action ActionToSync) Todo {
	newTodo := Todo{
		Title:     *action.Title,
		Completed: *action.Completed,
	}
	sql := `INSERT INTO todo_items(
  		title,
			completed
		) VALUES(
			$1,
			$2
		) RETURNING id;`
	err := model.db.QueryRow(sql, newTodo.Title, newTodo.Completed).Scan(&newTodo.Id)
	if err != nil {
		panic(fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err))
	}
	return newTodo
}

func (model *DbModel) UpdateDeviceActionToSyncIdToOutputJson(device Device) {
	actionToSyncIdToOutputJson, err :=
		json.Marshal(mapIntIntToMapStringInt(device.ActionToSyncIdToOutput))
	if err != nil {
		panic(fmt.Errorf(fmt.Sprintf("Error marshaling JSON: %s", err)))
	}

	sql := `UPDATE devices SET
		  action_to_sync_id_to_output_json = $1
			WHERE id = $2;`
	_, err = model.db.Exec(sql, string(actionToSyncIdToOutputJson), device.Id)
	if err != nil {
		panic(fmt.Errorf(`Error from db.Exec with sql=%s,
			  action_to_sync_id_to_output_json=%s, id=%d: %s`,
			sql, string(actionToSyncIdToOutputJson), device.Id, err))
	}
}

// returns number of rows updated (0 or 1)
func (model *DbModel) UpdateTodo(action ActionToSync, todoId int) int {
	setSqls := []string{}
	values := []interface{}{todoId} // first value is todoId
	if action.Completed != nil {
		setSqls = append(setSqls, fmt.Sprintf("completed = $%d", len(values)+1))
		values = append(values, action.Completed)
	}
	if action.Title != nil {
		setSqls = append(setSqls, fmt.Sprintf("title = $%d", len(values)+1))
		values = append(values, action.Title)
	}

	if len(values) > 0 {
		sql := "UPDATE todo_items SET " + strings.Join(setSqls, ", ") + " WHERE id = $1;"
		result, err := model.db.Exec(sql, values...)
		if err != nil {
			panic(fmt.Errorf(`Error from db.Exec with sql=%s, values=%v, id=%d: %s`,
				sql, values, todoId, err))
		}
		return convertRowsAffectedToInt(result.RowsAffected())
	} else {
		return 0
	}
}

func (model *DbModel) ListTodos() []Todo {
	sql := `SELECT id, title, completed FROM todo_items;`
	rows, err := model.db.Query(sql)
	if err != nil {
		panic(fmt.Sprintf("Error from db.Query with sql=%s: %s", sql, err))
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.Id, &todo.Title, &todo.Completed); err != nil {
			panic(fmt.Sprintf("Error from rows.Scan: %s", err))
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("Error from rows.Err: %s", err))
	}
	return todos
}

func (model *DbModel) DeleteTodo(todoId int) int {
	sql := `DELETE FROM todo_items WHERE id = $1;`
	result, err := model.db.Exec(sql, todoId)
	if err != nil {
		panic(fmt.Errorf(`Error from db.Exec with sql=%s, todoId=%d: %s`,
			sql, todoId, err))
	}
	return convertRowsAffectedToInt(result.RowsAffected())
}

func mapIntIntToMapStringInt(input map[int]int) map[string]int {
	output := map[string]int{}
	for k, v := range input {
		output[strconv.Itoa(k)] = v
	}
	return output
}

func mapStringIntToMapIntInt(input map[string]int) map[int]int {
	output := map[int]int{}
	for kString, v := range input {
		kInt, err := strconv.Atoi(kString)
		if err != nil {
			panic(fmt.Errorf("Error from strconv.Atoi for '%s': %s", kString, err))
		}
		output[kInt] = v
	}
	return output
}

func convertRowsAffectedToInt(i int64, err error) int {
	if err != nil {
		panic(fmt.Errorf("Error from RowsAffected(): %s", err))
	}

	if int64(int(i)) == i { // if round-trip conversion succeeds
		return int(i)
	} else {
		panic(fmt.Errorf("Couldn't convert int64 %d to int", i))
	}
}
