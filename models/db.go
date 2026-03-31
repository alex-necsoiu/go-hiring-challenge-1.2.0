package models

import "gorm.io/gorm"

// DBInterface defines the minimal interface for database operations needed by repositories.
// This interface allows for mocking and testing without requiring actual gorm.DB instances.
type DBInterface interface {
	Preload(column string, conditions ...interface{}) DBInterface
	Find(dest interface{}) DBInterface
	Joins(query string, args ...interface{}) DBInterface
	Where(query interface{}, args ...interface{}) DBInterface
	Model(value interface{}) DBInterface
	Count(count *int64) DBInterface
	Offset(offset int) DBInterface
	Limit(limit int) DBInterface
	First(dest interface{}) DBInterface
	Create(value interface{}) DBInterface
	GetError() error
}

// GormDBAdapter wraps a *gorm.DB and implements DBInterface.
// The adapter methods are thin pass-through wrappers with no business logic.
// They are covered through integration tests via the repositories that use them.
type GormDBAdapter struct {
	db *gorm.DB
}

// NewGormDBAdapter creates a new adapter for *gorm.DB
func NewGormDBAdapter(db *gorm.DB) DBInterface {
	return &GormDBAdapter{db: db}
}

func (a *GormDBAdapter) Preload(column string, conditions ...interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.Preload(column, conditions...)}
}

func (a *GormDBAdapter) Find(dest interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.Find(dest)}
}

func (a *GormDBAdapter) Joins(query string, args ...interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.Joins(query, args...)}
}

func (a *GormDBAdapter) Where(query interface{}, args ...interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.Where(query, args...)}
}

func (a *GormDBAdapter) Model(value interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.Model(value)}
}

func (a *GormDBAdapter) Count(count *int64) DBInterface {
	return &GormDBAdapter{db: a.db.Count(count)}
}

func (a *GormDBAdapter) Offset(offset int) DBInterface {
	return &GormDBAdapter{db: a.db.Offset(offset)}
}

func (a *GormDBAdapter) Limit(limit int) DBInterface {
	return &GormDBAdapter{db: a.db.Limit(limit)}
}

func (a *GormDBAdapter) First(dest interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.First(dest)}
}

func (a *GormDBAdapter) Create(value interface{}) DBInterface {
	return &GormDBAdapter{db: a.db.Create(value)}
}

func (a *GormDBAdapter) GetError() error {
	return a.db.Error
}
