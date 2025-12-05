//go:build linux && amd64

package config

import (
	"fmt"
	"time"

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

// --- JobData setters ---

func SetJobQueuedRemotely(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.QueuedRemotely, v)
	})
}

func SetJobMode(v *JobMode) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.Mode, v)
	})
}

func SetSelectedDisk(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.SelectedDisk, v)
	})
}

func SetEraseQueued(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.EraseQueued, v)
	})
}

func SetEraseMode(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.EraseMode, v)
	})
}

func SetSecureEraseCapable(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.SecureEraseCapable, v)
	})
}

func SetUsedSecureErase(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.UsedSecureErase, v)
	})
}

func SetEraseVerified(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.EraseVerified, v)
	})
}

func SetEraseVerifyPercent(v *float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.EraseVerifyPcnt, v)
	})
}

func SetEraseCompleted(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.EraseCompleted, v)
	})
}

func SetCloneQueued(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.CloneQueued, v)
	})
}

func SetCloneMode(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.CloneMode, v)
	})
}

func SetCloneImageName(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.CloneImageName, v)
	})
}

func SetCloneSourceHost(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.CloneSourceHost, v)
	})
}

func SetCloneCompleted(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.CloneCompleted, v)
	})
}

func SetJobFailed(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.Failed, v)
	})
}

func SetJobFailedMessage(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.FailedMessage, v)
	})
}

func SetJobStartTime(t *time.Time) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.StartTime, t)
	})
}

func SetJobEndTime(t *time.Time) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.EndTime, t)
	})
}

func SetJobDuration(d time.Duration) {
	if d.Seconds() < 0 {
		return
	}
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Duration == d {
			return false
		}
		cd.JobData.Duration = d
		return true
	})
}

func SetJobHibernated(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		return updateOptional(&cd.JobData.Hibernated, v)
	})
}
