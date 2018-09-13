package db

import (
	"testing"
)

func TestGetConnection(t *testing.T) {
	connection := GetConnection()
	if connection == nil {
		t.Error("could not connect to database.")
	}
	defer connection.Close()
}
