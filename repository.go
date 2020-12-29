package gormrepository

import (
	"fmt"

	"github.com/aklinkert/go-logging"
	"gorm.io/gorm"
)

type gormRepository struct {
	logger       logging.Logger
	db           *gorm.DB
	defaultJoins []string
}

// NewGormRepository returns a new base repository that implements TransactionRepository
func NewGormRepository(db *gorm.DB, logger logging.Logger, defaultJoins ...string) TransactionRepository {
	return &gormRepository{
		defaultJoins: defaultJoins,
		logger:       logger,
		db:           db,
	}
}

func (r *gormRepository) DB() *gorm.DB {
	return r.DBWithPreloads(nil)
}

func (r *gormRepository) GetAll(target interface{}, preloads ...string) error {
	r.logger.Debugf("Executing GetAll on %T", target)

	res := r.DBWithPreloads(preloads).Unscoped().Find(target)
	return r.HandleError(res)
}

func (r *gormRepository) GetByField(target interface{}, field string, value interface{}, preloads ...string) error {
	r.logger.Debugf("Executing GetByField on %T with %v = %v", target, field, value)

	res := r.DBWithPreloads(preloads).Where(fmt.Sprintf("%v = ?", field), value).Find(target)
	return r.HandleError(res)
}

func (r *gormRepository) GetByFields(target interface{}, filters map[string]interface{}, preloads ...string) error {
	r.logger.Debugf("Executing GetByField on %T with filters = %+v", target, filters)

	db := r.DBWithPreloads(preloads)
	for field, value := range filters {
		db = db.Where(fmt.Sprintf("%v = ?", field), value)
	}

	res := db.Find(target)
	return r.HandleError(res)
}

func (r *gormRepository) GetOneByField(target interface{}, field string, value interface{}, preloads ...string) error {
	r.logger.Debugf("Executing GetOneByField on %T with %v = %v", target, field, value)

	res := r.DBWithPreloads(preloads).Where(fmt.Sprintf("%v = ?", field), value).First(target)
	return r.HandleOneError(res)
}

func (r *gormRepository) GetOneByFields(target interface{}, filters map[string]interface{}, preloads ...string) error {
	r.logger.Debugf("Executing FindOneByField on %T with filters = %+v", target, filters)

	db := r.DBWithPreloads(preloads)
	for field, value := range filters {
		db = db.Where(fmt.Sprintf("%v = ?", field), value)
	}

	res := db.First(target)
	return r.HandleOneError(res)
}

func (r *gormRepository) GetOneByID(target interface{}, id string, preloads ...string) error {
	r.logger.Debugf("Executing GetOneByID on %T with ID %v", target, id)

	res := r.DBWithPreloads(preloads).Where("id = ?", id).First(target)
	return r.HandleOneError(res)
}

func (r *gormRepository) Create(target interface{}) error {
	r.logger.Debugf("Executing Create on %T", target)

	res := r.db.Create(target)
	return r.HandleError(res)
}

func (r *gormRepository) CreateTx(target interface{}, tx *gorm.DB) error {
	r.logger.Debugf("Executing Create on %T", target)

	res := tx.Create(target)
	return r.HandleError(res)
}

func (r *gormRepository) Save(target interface{}) error {
	r.logger.Debugf("Executing Save on %T", target)

	res := r.db.Save(target)
	return r.HandleError(res)
}

func (r *gormRepository) SaveTx(target interface{}, tx *gorm.DB) error {
	r.logger.Debugf("Executing Save on %T", target)

	res := tx.Save(target)
	return r.HandleError(res)
}

func (r *gormRepository) Delete(target interface{}) error {
	r.logger.Debugf("Executing Delete on %T", target)

	res := r.db.Delete(target)
	return r.HandleError(res)
}

func (r *gormRepository) DeleteTx(target interface{}, tx *gorm.DB) error {
	r.logger.Debugf("Executing Delete on %T", target)

	res := tx.Delete(target)
	return r.HandleError(res)
}

func (r *gormRepository) HandleError(res *gorm.DB) error {
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		err := fmt.Errorf("Error: %w", res.Error)
		r.logger.Error(err)
		return err
	}

	return nil
}

func (r *gormRepository) HandleOneError(res *gorm.DB) error {
	if err := r.HandleError(res); err != nil {
		return err
	}

	if res.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (r *gormRepository) DBWithPreloads(preloads []string) *gorm.DB {
	dbConn := r.db

	for _, join := range r.defaultJoins {
		dbConn = dbConn.Joins(join)
	}

	for _, preload := range preloads {
		dbConn = dbConn.Preload(preload)
	}

	return dbConn
}
