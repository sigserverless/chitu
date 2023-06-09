package utils

import "github.com/google/uuid"

func NewUId() string {
	// make uuid
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return id.String()
}
