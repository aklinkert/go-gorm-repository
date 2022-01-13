# go-gorm-repository

This is a simple repository implementation for [GORM](https://gorm.io/) providing basic functions to CRUD and query
entities as well as transactions and common error handling.

## Example

For a complete list of supported methods, please see [types.go](types.go).

```go
package main

import (
	"github.com/sirupsen/logrus"
	"github.com/aklinkert/go-gorm-repository"
	"gorm.io/gorm"
)

func main() {
	logger := logrus.New()
	db, _ := gorm.Open(_, _)
	
	// third parameter is a list of related entities that should always preload
	repo := gormrepository.NewGormRepository(db, logger, "CreatorUser", "Organization")
	
	instance := &exampleModel{}

	if err := repo.Create(instance); err != nil {
		logger.Fatalf("failed to create cache instance: %v", err)
	}
}
```

If you want to extend the repository with your own base or model repository just embed the interface into your implementation like so:

```go
package base

import "github.com/aklinkert/go-gorm-repository"

type BaseRepository interface {
	gormrepository.TransactionRepository
	GetOneByName(target interface{}, name string, preloads ...string) error
}

type repository struct {
	gormrepository.TransactionRepository
}

func NewRepository(db *gorm.DB, logger logging.Logger) BaseRepository {
	return &repository{
		TransactionRepository: gormrepository.NewGormRepository(db, logger, "Creator"),
	}
}

func (r *repository) GetOneByName(target interface{}, name string, preloads ...string) error {
	return r.TransactionRepository.GetOneByField(target, "name", name, preloads...)
}
```

## License

    Apache 2.0 License
