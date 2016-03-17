package models

import (
	"database/sql"
	"fmt"
	"strings"
)

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
	return err
}

func (model *DbModel) restartSequence(sequenceName string) error {
	sql := fmt.Sprintf("ALTER SEQUENCE \"%s\" RESTART WITH 1;", sequenceName)
	_, err := model.db.Exec(sql)
	return err
}

func (model *DbModel) FindOrCreateDeviceByUid(uid string) (*Device, error) {
	var device Device
	findSql := `SELECT id, uid
		FROM devices
		WHERE uid = $1`
	find1Err := model.db.QueryRow(findSql, uid).Scan(&device.Id, &device.Uid)

	if find1Err == nil {
		return &device, nil
	} else if find1Err == sql.ErrNoRows {
		insertSql := `INSERT INTO devices(
			uid,
			action_to_sync_id_to_output_json,
			completed_action_to_sync_id
		) VALUES(
			$1,
			'{}',
			0
		) RETURNING uid;`
		_, insertErr := model.db.Exec(insertSql, uid)
		if insertErr == nil {
			find2Err := model.db.QueryRow(findSql, uid).Scan(&device.Id, &device.Uid)
			if find2Err == nil {
				return &device, nil
			} else {
				return nil, fmt.Errorf(
					"Error from db.QueryRow with findSql=%s uid=%s find2Err=%s",
					findSql, uid, find2Err)
			}
		} else {
			if strings.HasPrefix(insertErr.Error(),
				"pq: duplicate key value violates unique constraint") {
				find2Err := model.db.QueryRow(findSql, uid).Scan(&device.Id, &device.Uid)
				if find2Err == nil {
					return &device, nil
				} else {
					return nil, fmt.Errorf(
						`Error from db.QueryRow after constraint failure
						with findSql=%s uid=%s findSql=%s`,
						findSql, uid, find2Err)
				}
			} else {
				return nil, fmt.Errorf(
					"Error from db.Exec with insertSql=%s uid=%s insertErr=%s",
					insertSql, uid, insertErr)
			}
		}
	} else {
		return nil, fmt.Errorf(
			"Error from db.QueryRows with findSql=%s uid=%s find1Err=%s",
			findSql, uid, find1Err)
	}
}
