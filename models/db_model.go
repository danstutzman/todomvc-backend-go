package models

import (
	"database/sql"
	"fmt"
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
	err := model.deleteFrom("devices")
	if err != nil {
		return fmt.Errorf("Error from deleteFrom(devices): %s", err)
	}

	err = model.restartSequence("devices_id_seq")
	if err != nil {
		return fmt.Errorf("Error from restartSequence(devices_id_seq): %s", err)
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

func (model *DbModel) FindOrCreateDeviceByUid(uid string) (*Device, error) {
	device, find1Err := model.findDeviceByUid(uid)
	if find1Err != nil {
		return nil, fmt.Errorf("Error from findDeviceByUid: %s", find1Err)
	}

	if device != nil {
		return device, nil
	} else {
		insertErr := model.createDevice(uid)
		if insertErr == nil {
			device, find2Err := model.findDeviceByUid(uid)
			if find2Err == nil {
				return device, nil
			} else {
				return nil, fmt.Errorf("Error from findDeviceByUid: %s", find2Err)
			}
		} else {
			if strings.HasPrefix(insertErr.Error(),
				"pq: duplicate key value violates unique constraint") {
				device, find2Err := model.findDeviceByUid(uid)
				if find2Err == nil {
					return device, nil
				} else {
					return nil, fmt.Errorf("Error from findDeviceByUid: %s", find2Err)
				}
			} else {
				return nil, fmt.Errorf("Error from createDevice: %s", uid, insertErr)
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

func (model *DbModel) findDeviceByUid(uid string) (*Device, error) {
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
			return nil, fmt.Errorf("Error from unmarshaling JSON '%s': %s",
				actionToSyncIdToOutputJson, err)
		}

		device.ActionToSyncIdToOutput, err =
			mapStringIntToMapIntInt(actionToSyncIdToOutput)
		if err != nil {
			return nil, fmt.Errorf("Error from mapStringIntToMapIntInt: %s", err)
		}

		return &device, nil
	} else if err == SqlErrNoRows {
		return nil, nil
	} else {
		return nil, fmt.Errorf("Error from db.QueryRow with sql=%s: %s", sql)
	}
}

func (model *DbModel) CreateTodo(action ActionToSync) (*Todo, error) {
	newTodo := Todo{
		Title:     action.Title,
		Completed: action.Completed,
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
		return nil, fmt.Errorf("Error from db.Exec with sql=%s: %s", sql, err)
	}
	return &newTodo, nil
}

func (model *DbModel) UpdateDeviceActionToSyncIdToOutputJson(device *Device) error {
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
