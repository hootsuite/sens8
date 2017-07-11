package sqlite3

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/mattes/migrate/driver/sqlite3"
	"github.com/mattes/migrate/migrate"
	_ "github.com/mattn/go-sqlite3"
)

func BuildDatabase(
	schemaPath *string,
	migrationsPath *string,
) (*gorm.DB, error) {
	f, _ := ioutil.TempFile("/tmp", "sqlite")
	db, err := gorm.Open("sqlite3", f.Name())

	if err != nil {
		return nil, err
	}

	if schemaPath != nil {
		blob, err := ioutil.ReadFile(*schemaPath)

		if err != nil {
			return nil, err
		}

		if _, err := db.DB().Exec(string(blob)); err != nil {
			return nil, err
		}
	}

	if migrationsPath != nil {
		dbPath := fmt.Sprintf("sqlite3://%s", f.Name())
		errs, ok := migrate.UpSync(dbPath, *migrationsPath)

		if !ok {
			strErrs := []string{}
			for _, migrationError := range errs {
				strErrs = append(strErrs, migrationError.Error())
			}
			return nil, errors.New(strings.Join(strErrs, ","))
		}
	}

	return db, nil
}
