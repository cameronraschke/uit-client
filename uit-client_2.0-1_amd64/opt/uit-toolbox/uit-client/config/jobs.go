//go:build linux && amd64

package config

import (
	"fmt"
	"time"
	"uitclient/types"

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
	if uid == uuid.Nil {
		return
	}
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.UUID == uid {
			return false
		}
		cd.JobData.UUID = uid
		return true
	})
}

func GetJobUUID() uuid.UUID {
	cd := GetClientData()
	if cd.JobData == nil {
		return uuid.Nil
	}
	val := cd.JobData.UUID
	return val
}

func SetJobQueuedRemotely(isQueuedRemotely *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.JobData.QueuedRemotely, isQueuedRemotely)
	})
}

func SetJobType(jobType *types.JobType) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.Type, jobType)
	})
}

func GetJobType() *types.JobType {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.Type
	return &val
}

func SetSelectedDisk(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.JobData.SelectedDisk, v)
	})
}

func GetSelectedDisk() *string {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.SelectedDisk
	return &val
}

func SetEraseQueued(eraseIsQueued *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.JobData.EraseQueued, eraseIsQueued)
	})
}

func GetEraseQueued() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.EraseQueued
	return &val
}

func SetEraseMode(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.JobData.EraseMode, v)
	})
}

func GetEraseMode() *string {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.EraseMode
	return &val
}

func SetSecureEraseCapable(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		return updateOptional(&cd.JobData.SecureEraseCapable, v)
	})
}

func GetSecureEraseCapable() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.SecureEraseCapable
	return &val
}

func SetUsedSecureErase(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.UsedSecureErase, v)
	})
}

func GetUsedSecureErase() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.UsedSecureErase
	return &val
}

func SetEraseVerified(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.EraseVerified, v)
	})
}

func GetEraseVerified() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.EraseVerified
	return &val
}

func SetEraseVerifyPercent(v *float64) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.EraseVerifyPcnt, v)
	})
}

func GetEraseVerifyPercent() *float64 {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.EraseVerifyPcnt
	return &val
}

func SetEraseCompleted(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.EraseCompleted, v)
	})
}

func GetEraseCompleted() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.EraseCompleted
	return &val
}

func SetCloneQueued(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.CloneQueued, v)
	})
}

func GetCloneQueued() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.CloneQueued
	return &val
}

func SetCloneMode(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.CloneMode, v)
	})
}

func GetCloneMode() *string {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.CloneMode
	return &val
}

func SetCloneImageName(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.CloneImageName, v)
	})
}

func GetCloneImageName() *string {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.CloneImageName
	return &val
}

func SetCloneSourceHost(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.CloneSourceHost, v)
	})
}

func GetCloneSourceHost() *string {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.CloneSourceHost
	return &val
}

func SetCloneCompleted(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.CloneCompleted, v)
	})
}

func GetCloneCompleted() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.CloneCompleted
	return &val
}

func SetJobFailed(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.Failed, v)
	})
}

func GetJobFailed() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.Failed
	return &val
}

func SetJobFailedMessage(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.FailedMessage, v)
	})
}

func GetJobFailedMessage() *string {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.FailedMessage
	return &val
}

func SetJobStartTime(t *time.Time) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.StartTime, t)
	})
}

func GetJobStartTime() *time.Time {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.StartTime
	return &val
}

func SetJobEndTime(t *time.Time) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.EndTime, t)
	})
}

func GetJobEndTime() *time.Time {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.EndTime
	return &val
}

func SetJobDuration(d time.Duration) {
	if d.Seconds() < 0 {
		return
	}
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Duration == d {
			return false
		}
		cd.JobData.Duration = d
		return true
	})
}

func GetJobDuration() *time.Duration {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := cd.JobData.Duration
	return &val
}

func SetJobHibernated(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		return updateOptional(&cd.JobData.Hibernated, v)
	})
}

func GetJobHibernated() *bool {
	cd := GetClientData()
	if cd.JobData == nil {
		return nil
	}
	val := *cd.JobData.Hibernated
	return &val
}
