package gormrepository

import (
	"gorm.io/gorm"
)

var (
	// ErrNotFound is a convenience reference for the actual GORM error
	ErrNotFound = gorm.ErrRecordNotFound
)
