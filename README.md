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


## License

    Apache 2.0 License
