package postgres

import (
	"fmt"
	"io/ioutil"
	"testing"

	_ "github.com/lib/pq"
)

func TestBuildDatabaseNotExist(t *testing.T) {
	dbPath := "/foo/bar"
	if _, err := BuildDatabase(&dbPath, nil); err == nil {
		t.Errorf("Wrong file")
	}
}

func TestBuildDatabaseNotValid(t *testing.T) {
	f, _ := ioutil.TempFile("/tmp", "fo")
	fName := f.Name()

	f.WriteString("foo;\nbar;")

	if _, err := BuildDatabase(&fName, nil); err == nil {
		t.Errorf("Execute wrong command")
	}
}

func TestBuildDatabaseValid(t *testing.T) {
	f, _ := ioutil.TempFile("/tmp", "fo")
	fName := f.Name()

	f.WriteString(
		`
		CREATE TABLE x(x INTEGER, y INTEGER, z INTEGER);
		CREATE TABLE y(x INTEGER, y INTEGER, z INTEGER);
		`,
	)

	db, err := BuildDatabase(&fName, nil)
	defer db.Close()
	if err != nil {
		t.Errorf("Cannot execute sql command: %s", err.Error())
	}

	r := -1

	db.DB().QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_catalog='test_database' AND table_schema='public';",
	).Scan(&r)

	if r != 2 {
		t.Errorf("Wrong number of table: %v", r)
	}
}

func TestBuildDatabaseValidFromMigration(t *testing.T) {
	f, _ := ioutil.TempFile("/tmp", "fo")
	fName := f.Name()
	fixturesPath := "../../fixtures"

	f.WriteString(
		`
		CREATE TABLE x(x INTEGER, y INTEGER, z INTEGER);
		CREATE TABLE y(x INTEGER, y INTEGER, z INTEGER);
		`,
	)

	db, err := BuildDatabase(&fName, &fixturesPath)
	defer db.Close()

	if err != nil {
		t.Errorf("Cannot execute sql command: %s", err.Error())
	}

	for _, l := range []string{"x", "y", "z"} {
		r := -1

		db.DB().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", l)).Scan(&r)

		if r != 0 {
			t.Errorf("Wrong number of table: %v", r)
		}
	}
}

func TestBuildDatabaseValidFromOnlyMigration(t *testing.T) {
	fixturesPath := "../../fixtures"

	db, err := BuildDatabase(nil, &fixturesPath)
	defer db.Close()

	if err != nil {
		t.Errorf("Cannot execute sql command: %s", err.Error())
	}

	for _, l := range []string{"x", "y"} {
		r := -1

		db.DB().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", l)).Scan(&r)

		if r != -1 {
			t.Errorf("Wrong number of table: %v", r)
		}
	}

	r := -1

	db.DB().QueryRow("SELECT COUNT(*) FROM z").Scan(&r)

	if r != 0 {
		t.Errorf("Wrong number of table: %v", r)
	}
}
