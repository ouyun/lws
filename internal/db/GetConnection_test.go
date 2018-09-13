package db

import (
	"testing"
)

func TestGetConnection(t *testing.T) {
	connection := GetConnection()
	if connection == nil {
		t.Errorf("could not connect to database: %v", err)
	}
	defer connection.Close()
}
