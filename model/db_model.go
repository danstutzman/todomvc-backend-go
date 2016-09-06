package model

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

func (model *DbModel) Reset() error {
	var err error

	err = model.deleteFrom("devices")
	if err != nil {
		return fmt.Errorf("Error from deleteFrom(devices): %s", err)
	}

	err = model.restartSequence("devices_id_seq")
	if err != nil {
		return fmt.Errorf("Error from restartSequence(devices_id_seq): %s", err)
	}

	err = model.deleteFrom("todo_items")
	if err != nil {
		return fmt.Errorf("Error from deleteFrom(todo_items): %s", err)
	}

	err = model.restartSequence("todo_items_id_seq")
	if err != nil {
		return fmt.Errorf("Error from restartSequence(todo_items_id_seq): %s", err)
	}

	return nil
}

func (model *DbModel) deleteFrom(tableName string) error {
	sql := fmt.Sprintf("DELETE FROM \"%s\"", tableName)
	_, err := model.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err)
	}
	return nil
}

func (model *DbModel) restartSequence(sequenceName string) error {
	sql := fmt.Sprintf("ALTER SEQUENCE \"%s\" RESTART WITH 1;", sequenceName)
	_, err := model.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err)
	}
	return nil
}

func (model *DbModel) FindOrCreateDeviceByUid(uid string) (Device, error) {
	device, find1Err := model.findDeviceByUid(uid)
	if find1Err != nil {
		return Device{}, fmt.Errorf("Error from findDeviceByUid: %s", find1Err)
	}

	if device.Id != 0 {
		return device, nil
	} else {
		insertErr := model.createDevice(uid)
		if insertErr == nil {
			device, find2Err := model.findDeviceByUid(uid)
			if find2Err == nil {
				return device, nil
			} else {
				return Device{}, fmt.Errorf("Error from findDeviceByUid: %s", find2Err)
			}
		} else {
			if strings.HasPrefix(insertErr.Error(),
				"pq: duplicate key value violates unique constraint") {
				device, find2Err := model.findDeviceByUid(uid)
				if find2Err == nil {
					return device, nil
				} else {
					return Device{}, fmt.Errorf("Error from findDeviceByUid: %s", find2Err)
				}
			} else {
				return Device{}, fmt.Errorf("Error from createDevice: %s", uid, insertErr)
			}
		}
	}
}

func (model *DbModel) createDevice(uid string) error {
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
		return fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err)
	}
	return err
}

func (model *DbModel) findDeviceByUid(uid string) (Device, error) {
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
			return Device{}, fmt.Errorf("Error from unmarshaling JSON '%s': %s",
				actionToSyncIdToOutputJson, err)
		}

		device.ActionToSyncIdToOutput, err =
			mapStringIntToMapIntInt(actionToSyncIdToOutput)
		if err != nil {
			return Device{}, fmt.Errorf("Error from mapStringIntToMapIntInt: %s", err)
		}

		return device, nil
	} else if err == SqlErrNoRows {
		return Device{}, nil
	} else {
		return Device{}, fmt.Errorf("Error from db.QueryRow with sql=%s: %s", sql)
	}
}

func (model *DbModel) CreateTodo(action ActionToSync) (Todo, error) {
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
		return Todo{}, fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err)
	}
	return newTodo, nil
}

func (model *DbModel) UpdateDeviceActionToSyncIdToOutputJson(device Device) error {
	actionToSyncIdToOutputJson, err :=
		json.Marshal(mapIntIntToMapStringInt(device.ActionToSyncIdToOutput))
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Error marshaling JSON: %s", err))
	}

	sql := `UPDATE devices SET
		  action_to_sync_id_to_output_json = $1
			WHERE id = $2;`
	_, err = model.db.Exec(sql, string(actionToSyncIdToOutputJson), device.Id)
	if err != nil {
		return fmt.Errorf(`Error from db.Exec with sql=%s,
			  action_to_sync_id_to_output_json=%s, id=%s: %s`,
			sql, string(actionToSyncIdToOutputJson), device.Id, err)
	}
	return nil
}

// returns number of rows updated (0 or 1)
func (model *DbModel) UpdateTodo(action ActionToSync, todoId int) (int, error) {
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
			return 0, fmt.Errorf(`Error from db.Exec with sql=%s, values=%v, id=%s: %s`,
				sql, values, todoId, err)
		}
		return int64ErrToIntErr(result.RowsAffected())
	} else {
		return 0, nil
	}
}

func (model *DbModel) ListTodos() ([]Todo, error) {
	sql := `SELECT id, title, completed FROM todo_items;`
	rows, err := model.db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("Error from db.Query with sql=%s", sql, err)
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.Id, &todo.Title, &todo.Completed); err != nil {
			return nil, fmt.Errorf("Error from rows.Scan: %s", err)
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error from rows.Err: %s", err)
	}
	return todos, nil
}

func (model *DbModel) DeleteTodo(todoId int) (int, error) {
	sql := `DELETE FROM todo_items WHERE id = $1;`
	result, err := model.db.Exec(sql, todoId)
	if err != nil {
		return 0, fmt.Errorf(`Error from db.Exec with sql=%s, todoId=%s: %s`,
			sql, todoId, err)
	}
	return int64ErrToIntErr(result.RowsAffected())
}

func mapIntIntToMapStringInt(input map[int]int) map[string]int {
	output := map[string]int{}
	for k, v := range input {
		output[strconv.Itoa(k)] = v
	}
	return output
}

func mapStringIntToMapIntInt(input map[string]int) (map[int]int, error) {
	output := map[int]int{}
	for kString, v := range input {
		kInt, err := strconv.Atoi(kString)
		if err != nil {
			return nil, fmt.Errorf("Error from strconv.Atoi for '%s': %s", kString, err)
		}
		output[kInt] = v
	}
	return output, nil
}

func int64ErrToIntErr(i int64, e error) (int, error) {
	if int64(int(i)) == i { // if round-trip conversion succeeds
		return int(i), e
	} else {
		return 0, fmt.Errorf("Couldn't convert int64 %d to int, also e=%s", i, e)
	}
}
