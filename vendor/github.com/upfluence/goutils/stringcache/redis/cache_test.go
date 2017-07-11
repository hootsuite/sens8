package redis

import (
	"testing"

	"github.com/upfluence/goutils/stringcache/testutil"
)

func TestIntegration(t *testing.T) {
	c, err := NewCache("redis://localhost/0", "test-key")

	if err != nil {
		t.Errorf("Error returned by the constructor %+v", err)
	}

	c.conn.Do("SDEL", c.key)

	testutil.IntegrationScenario(t, c)
}
