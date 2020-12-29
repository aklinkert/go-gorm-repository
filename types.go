package gormrepository

import (
	"gorm.io/gorm"
)

// Repository is a generic DB handler that cares about default error handling
type Repository interface {
	GetAll(target interface{}, preloads ...string) error
	GetWhere(target interface{}, condition string, preloads ...string) error
	GetByField(target interface{}, field string, value interface{}, preloads ...string) error
	GetByFields(target interface{}, filters map[string]interface{}, preloads ...string) error

	GetOneByField(target interface{}, field string, value interface{}, preloads ...string) error
	GetOneByFields(target interface{}, filters map[string]interface{}, preloads ...string) error
	GetOneByID(target interface{}, id string, preloads ...string) error

	Create(target interface{}) error
	Save(target interface{}) error
	Delete(target interface{}) error

	DB() *gorm.DB
	DBWithPreloads(preloads []string) *gorm.DB
	HandleError(res *gorm.DB) error
	HandleOneError(res *gorm.DB) error
}

// TransactionRepository extends Repository with modifier functions that accept a transaction
type TransactionRepository interface {
	Repository
	CreateTx(target interface{}, tx *gorm.DB) error
	SaveTx(target interface{}, tx *gorm.DB) error
	DeleteTx(target interface{}, tx *gorm.DB) error
}
