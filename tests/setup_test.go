package tests

import (
	"github.com/ldchengyi/linkflow/internal/testutil"
)

var testCfg *testutil.TestConfig

func init() {
	cfg, err := testutil.LoadTestConfig()
	if err != nil {
		panic("load test config: " + err.Error())
	}
	testCfg = cfg
}
