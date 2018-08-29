package db

import (
	"testing"
)

func TestConnect(t *testing.T) {
	connection, err := Connect()
	if err != nil {
		t.Errorf("could not connect to database: %v", err)
	}
	defer connection.Close()
}
