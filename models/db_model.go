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
	sql := `SELECT id, uid
		FROM devices
		WHERE uid = $1`
	err := model.db.QueryRow(sql, uid).Scan(&device.Id, &device.Uid)
	device.ActionToSyncIdToOutput = map[int]int{}
	if err == nil {
		return &device, nil
	} else if err == SqlErrNoRows {
		return nil, nil
	} else {
		return nil, fmt.Errorf("Error from db.QueryRow with sql=%s: %s", sql)
	}
}
