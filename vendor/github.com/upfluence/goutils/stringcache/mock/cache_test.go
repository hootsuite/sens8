package mock

import (
	"testing"

	"github.com/upfluence/goutils/stringcache/testutil"
)

func TestIntegration(t *testing.T) {
	testutil.IntegrationScenario(t, NewCache())
}
