package config

import (
	"fmt"

	"github.com/google/uuid"
)

func CreateNewJobUUID() (uuid.UUID, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error generating new UUID: %v", err)
	}
	return newUUID, nil
}

func SetJobUUID(uid uuid.UUID) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.UUID == uid {
			return false
		}
		cd.UUID = uid
		return true
	})
}

func GetJobUUID() uuid.UUID {
	cd := GetClientData()
	if cd == nil {
		return uuid.Nil
	}
	return cd.UUID
}
