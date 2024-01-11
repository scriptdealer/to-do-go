package storage

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("DB_HOST", "test")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_PASS", "test")
	getConfig()
}
