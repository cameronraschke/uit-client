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

func UpdateJobData(mutate func(*JobData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		mutate(cd.JobData) // mutate in place, already isolated by copy-on-write
		return true
	})
}
