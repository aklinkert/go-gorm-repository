# go-gorm-repository

[![test](https://github.com/aklinkert/go-gorm-repository/actions/workflows/test.yml/badge.svg)](https://github.com/aklinkert/go-gorm-repository/actions/workflows/test.yml)

This is a simple repository implementation for [GORM](https://gorm.io/) providing basic functions to CRUD and query
entities as well as transactions and common error handling.

## Installation

```sh
go get github.com/aklinkert/go-gorm-repository
```

Requires Go 1.21 or newer.

## Example

For a complete list of supported methods, please see [types.go](types.go). The logger is a
[`logging.Logger`](https://github.com/aklinkert/go-logging), which a logrus logger or entry satisfies.

```go
package main

import (
	gormrepository "github.com/aklinkert/go-gorm-repository"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	logger := logrus.New()

	// initialize your DB Connection
	var dial gorm.Dialector

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		logger.Fatalf("failed to open DB connection: %v", err)
	}

	// the remaining parameters are relations that are always joined
	repo := gormrepository.NewGormRepository(db, logger, "CreatorUser", "Organization")

	instance := &exampleModel{}

	if err := repo.Create(instance); err != nil {
		logger.Fatalf("failed to create instance: %v", err)
	}
}
```

The `GetOne*` methods return `gormrepository.ErrNotFound` when no record matches, the list methods return an empty
result instead.

The field names passed to `GetByField`, `GetByFields`, `GetOneByField`, `GetOneByFields` and their batch variants are
used as given. With default joins, a column that exists in both tables is ambiguous, so qualify it with the table name,
e.g. `repo.GetByField(&posts, "posts.title", "Hello")`.

`GetWhere` conditions and field names end up in the SQL as they are, so never build them from user input.
Values are always passed as bound parameters.

If you want to extend the repository with your own base or model repository just embed the interface into your
implementation like so:

```go
package base

import (
	"github.com/aklinkert/go-logging"
	gormrepository "github.com/aklinkert/go-gorm-repository"
	"gorm.io/gorm"
)

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

Apache 2.0, see [LICENSE](LICENSE).
