package coreclient

import (
	"github.com/satori/go.uuid"
)

func generateUuidString() string {
	id, err := uuid.NewV4()
	if err != nil {
		errorLogger("generate uuidv4 failed: [%s]", err)
		panic("generate uuidv4 failed")
	}
	return id.String()
}
