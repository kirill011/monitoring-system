package repo

import "errors"

var (
	ErrDeviceExists = errors.New("device already exists")
	ErrNotFound     = errors.New("device not found")
)
