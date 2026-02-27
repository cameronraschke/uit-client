//go:build linux && amd64

package config

import "uitclient/types"

// Only getters, so do http request
func GetJobNameFormatted(jobName *string) *string {
	if jobName == nil {
		return nil
	}
	return nil
}

func GetAvailableJobs(jobInQuestion *string) *bool {
	return nil // iterate through available jobs and check if jobInQuestion is available
}

// the remaining setters and getters
func SetJobName(jobName *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobName, jobName)
	})
}

func GetJobName() *string {
	cd := GetClientData()
	if cd.JobData == nil || cd.JobData.Realtime == nil || cd.JobData.Realtime.JobName == nil {
		return nil
	}
	v := *cd.JobData.Realtime.JobName
	return &v
}

func SetJobQueued(jobQueued *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueued, jobQueued)
	})
}

func GetJobQueued() *bool {
	cd := GetClientData()
	if cd.JobData == nil || cd.JobData.Realtime == nil || cd.JobData.Realtime.JobQueued == nil {
		return nil
	}
	v := *cd.JobData.Realtime.JobQueued
	return &v
}

func SetJobActive(jobActive *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobActive, jobActive)
	})
}

func GetJobActive() *bool {
	cd := GetClientData()
	if cd.JobData == nil || cd.JobData.Realtime == nil || cd.JobData.Realtime.JobActive == nil {
		return nil
	}
	v := *cd.JobData.Realtime.JobActive
	return &v
}

func SetJobQueuePosition(position *int) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueuePosition, position)
	})
}

func GetJobQueuePosition() *int {
	cd := GetClientData()
	if cd.JobData == nil || cd.JobData.Realtime == nil || cd.JobData.Realtime.JobQueuePosition == nil {
		return nil
	}
	v := *cd.JobData.Realtime.JobQueuePosition
	return &v
}
