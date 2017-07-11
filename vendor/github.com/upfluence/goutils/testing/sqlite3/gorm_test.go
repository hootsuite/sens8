package sqlite3

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var fixturesPath = "../../fixtures"

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
		CREATE TABLE x(x INTEGER PRIMARY KEY ASC, y, z);
		CREATE TABLE y(x INTEGER PRIMARY KEY ASC, y, z);
		`,
	)

	db, err := BuildDatabase(&fName, nil)

	if err != nil {
		t.Errorf("Cannot execute sql command")
	}

	r := -1

	db.DB().QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = \"table\";",
	).Scan(&r)

	if r != 2 {
		t.Errorf("Wrong number of table: %v", r)
	}
}

func TestBuildDatabaseValidFromMigration(t *testing.T) {
	f, _ := ioutil.TempFile("/tmp", "fo")
	fName := f.Name()

	f.WriteString(
		`
		CREATE TABLE x(x INTEGER PRIMARY KEY ASC, y, z);
		CREATE TABLE y(x INTEGER PRIMARY KEY ASC, y, z);
		`,
	)

	db, err := BuildDatabase(&fName, &fixturesPath)

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
	db, err := BuildDatabase(nil, &fixturesPath)

	if err != nil {
		t.Errorf("Cannot execute sql command")
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
