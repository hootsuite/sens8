package testutil

import (
	"testing"

	"github.com/upfluence/goutils/stringcache"
)

const mockKey = "foo-bar"

func IntegrationScenario(t *testing.T, cache stringcache.Cache) {
	e, err := cache.Has(mockKey)

	if e {
		t.Error("Key exist in empty cache")
	}

	if err != nil {
		t.Errorf("Has returned an error: %+v", err)
	}

	if err2 := cache.Delete(mockKey); err2 != stringcache.ErrNotFound {
		t.Errorf("Delete returned an unexpected error: %+v", err2)
	}

	if err2 := cache.Add(mockKey); err2 != nil {
		t.Errorf("Add returned an unexpected error: %+v", err2)
	}

	e, err = cache.Has(mockKey)

	if !e {
		t.Error("Key not exist in a warmed up cache")
	}

	if err != nil {
		t.Errorf("Has returned an error: %+v", err)
	}

	if err := cache.Delete(mockKey); err != nil {
		t.Errorf("Delete returned an unexpected error: %+v", err)
	}
}
