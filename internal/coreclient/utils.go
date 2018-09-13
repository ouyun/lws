package coreclient

import (
	"github.com/satori/go.uuid"
)

func generateUuidString() string {
	id := uuid.NewV4()
	return id.String()
}
