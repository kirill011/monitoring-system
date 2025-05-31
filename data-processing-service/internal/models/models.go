package models

import (
	"database/sql"
	"regexp"
	"time"
)

type Tag struct {
	ID            int32      `db:"id"`
	Name          string     `db:"name"`
	DeviceId      int32      `db:"device_id"`
	Regexp        string     `db:"regexp"`
	CompareType   string     `db:"compare_type"`
	Value         string     `db:"value"`
	ArrayIndex    int32      `db:"array_index"`
	Subject       string     `db:"subject"`
	SeverityLevel string     `db:"severity_level"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`

	CompareFunc    func(string, string) bool `json:"-" db:"-"`
	CompiledRegexp *regexp.Regexp            `json:"-" db:"-"`
}

func (t Tag) Compare(val string) bool {
	return t.CompareFunc(val, t.Value)
}

type Device struct {
	ID          int32
	Name        string
	DeviceType  string
	Address     string
	Responsible []int32
}

type Message struct {
	Id            int32     `db:"id"`
	GotAt         time.Time `db:"got_at"`
	DeviceId      int32     `db:"device_id"`
	Message       string    `db:"message"`
	MessageType   string    `db:"message_type"`
	SeverityLevel string    `db:"severity_level"`
	Component     string    `db:"component"`
}

type CountByDeviceID struct {
	DeviceId int32 `db:"device_id"`
	Count    int32 `db:"count"`
}

type SendedNotification struct {
	Message   string
	DeviceId  int32
	ExpiredAt time.Time
}

type MonthReportRow struct {
	DeviceID               int32           `db:"device_id"`
	MessageType            string          `db:"message_type"`
	ActiveDays             int32           `db:"active_days"`
	TotalMessages          int64           `db:"total_messages"`
	AvgDailyMessages       float64         `db:"avg_daily_messages"`
	MaxDailyMessages       int64           `db:"max_daily_messages"`
	MedianDailyMessages    float64         `db:"median_daily_messages"`
	TotalCritical          int64           `db:"total_critical"`
	MaxDailyCritical       int64           `db:"max_daily_critical"`
	MaxDailyComponents     int32           `db:"max_daily_components"`
	MostActiveComponent    sql.NullString  `db:"most_active_component"`
	FirstCriticalTime      sql.NullTime    `db:"first_critical_time"`
	LastCriticalTime       sql.NullTime    `db:"last_critical_time"`
	AvgCriticalIntervalSec sql.NullFloat64 `db:"avg_critical_interval_sec"`
	CriticalPercentage     float64         `db:"critical_percentage"`
	OverallVolumeRank      int32           `db:"overall_volume_rank"`
}
