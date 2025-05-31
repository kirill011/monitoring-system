package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Device struct {
	ID          int32            `db:"id"`
	Name        string           `db:"name"`
	DeviceType  string           `db:"device_type"`
	Address     string           `db:"address"`
	Responsible SqlJsonbIntArray `db:"responsible"`
	CreatedAt   *time.Time       `db:"created_at"`
	UpdatedAt   *time.Time       `db:"updated_at"`
}

type SqlJsonbIntArray []int32

func (arr SqlJsonbIntArray) Value() (driver.Value, error) {
	res, err := json.Marshal(arr)
	return res, err
}
func (arr *SqlJsonbIntArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		err := fmt.Errorf("SqlJsonbStringArray: item not created")
		return err
	}
	err := json.Unmarshal(b, &arr)
	return err
}
