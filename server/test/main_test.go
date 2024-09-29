package test

import (
	"os"
	"testing"

	"github.com/marcelhfm/home_server/test/setup"
)

func TestMain(m *testing.M) {
	testsetup.StartPostgresContainer(nil)

	exitCode := m.Run()

	testsetup.TearDownTests(nil)
	os.Exit(exitCode)
}
