package models

import "time"

type Device struct {
	ID          int32     `db:"id"`
	Name        string    `db:"name"`
	DeviceType  string    `db:"device_type"`
	Address     string    `db:"address"`
	Responsible []int32   `db:"responsible"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Message struct {
	Id          int32  `db:"id"`
	DeviceId    int32  `db:"device_id"`
	Message     string `db:"message"`
	MessageType string `db:"message_type"`
	Component   string `db:"component"`
	DeviceIP    string `db:"device_ip"`
}
