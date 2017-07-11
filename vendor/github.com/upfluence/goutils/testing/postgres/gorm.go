package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
)

const defaultPostgresURL = "postgres://localhost:5432/test_database?sslmode=disable"

func BuildDatabase(
	schemaPath *string,
	migrationsPath *string,
) (*gorm.DB, error) {
	postgresURL := os.Getenv("POSTGRES_URL")

	if postgresURL == "" {
		postgresURL = defaultPostgresURL
	}

	postgresURLSlipped := strings.Split(postgresURL, "/")

	database := strings.Split(
		postgresURLSlipped[len(postgresURLSlipped)-1],
		"?",
	)[0]

	plainURL := postgresURLSlipped[0 : len(postgresURLSlipped)-1]
	plainDB, err := sql.Open(
		"postgres",
		fmt.Sprintf("%s?sslmode=disable", strings.Join(plainURL, "/")),
	)

	if err != nil {
		return nil, err
	}

	plainDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", database))
	plainDB.Exec(fmt.Sprintf("CREATE DATABASE %s", database))

	plainDB.Close()

	db, err := gorm.Open("postgres", postgresURL)

	if err != nil {
		db.Close()
		return nil, err
	}

	if schemaPath != nil {
		blob, err := ioutil.ReadFile(*schemaPath)

		if err != nil {
			db.Close()
			return nil, err
		}

		if _, err := db.DB().Exec(string(blob)); err != nil {
			db.Close()
			return nil, err
		}
	}

	if migrationsPath != nil {
		errs, ok := migrate.UpSync(postgresURL, *migrationsPath)

		if !ok {
			strErrs := []string{}
			for _, migrationError := range errs {
				strErrs = append(strErrs, migrationError.Error())
			}

			db.Close()
			return nil, errors.New(strings.Join(strErrs, ","))
		}
	}

	return db, nil
}
