//go:build linux && amd64

package config

import "time"

func UpdateRealtimeJobQueue(mutate func(*JobQueueData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		mutate(cd.JobData.Realtime)
		return true
	})
}

func SetRealtimeJobName(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobName, v)
	})
}

func SetRealtimeJobNameFormatted(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobNameFormatted, v)
	})
}

func SetRealtimeJobQueued(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueued, v)
	})
}

func SetRealtimeJobRequiresQueue(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobRequiresQueue, v)
	})
}

func SetRealtimeJobQueuePosition(v *int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueuePosition, v)
	})
}

func SetRealtimeJobQueuedOverride(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueuedOverride, v)
	})
}

func SetRealtimeJobActive(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobActive, v)
	})
}

func SetRealtimeJobAvailable(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobAvailable, v)
	})
}

func SetRealtimeJobProgress(v *float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobProgress, v)
	})
}

func SetRealtimeJobDuration(d time.Duration) {
	if d.Seconds() < 0 {
		return
	}
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		if cd.JobData.Realtime.JobDuration == d {
			return false
		}
		cd.JobData.Realtime.JobDuration = d
		return true
	})
}

func SetRealtimeJobStatusMessage(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobStatusMessage, v)
	})
}

// UpdateRealtimeSystemData mutates the realtime system subtree using copy-on-write.
func UpdateRealtimeSystemData(mutate func(*RealtimeSystemData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		mutate(cd.RealtimeSystemData)
		return true
	})
}

// --- RealtimeSystemData setters ---

func SetLastHeardTimestamp(t *time.Time) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.LastHeardTimestamp, t)
	})
}

func SetBootTimestamp(t *time.Time) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.BootTimestamp, t)
	})
}

func SetCurrentTimestamp(t *time.Time) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.CurrentTimestamp, t)
	})
}

func SetUptime(d time.Duration) {
	if d.Seconds() < 0 {
		return
	}
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.Uptime == d {
			return false
		}
		cd.RealtimeSystemData.Uptime = d
		return true
	})
}

func SetKernelUpdated(v *bool) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.KernelUpdated, v)
	})
}

func SetBase64Screenshot(v *string) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.Base64Screenshot, v)
	})
}

// Nested subtree updaters (realtime hardware/software)

func UpdateRealtimeHardware(mutate func(*ClientHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.Hardware == nil {
			cd.RealtimeSystemData.Hardware = &ClientHardwareData{}
		}
		mutate(cd.RealtimeSystemData.Hardware)
		return true
	})
}

func UpdateRealtimeSoftware(mutate func(*ClientSoftwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.Software == nil {
			cd.RealtimeSystemData.Software = &ClientSoftwareData{}
		}
		mutate(cd.RealtimeSystemData.Software)
		return true
	})
}

func UpdateRealtimeResourceUsage(mutate func(*ClientResourceUsageData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.ResourceUsage == nil {
			cd.RealtimeSystemData.ResourceUsage = &ClientResourceUsageData{}
		}
		mutate(cd.RealtimeSystemData.ResourceUsage)
		return true
	})
}

// --- Realtime ResourceUsage setters ---

func SetRealtimeEnergyUsage(v *float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.ResourceUsage == nil {
			cd.RealtimeSystemData.ResourceUsage = &ClientResourceUsageData{}
		}
		return updateOptional(&cd.RealtimeSystemData.ResourceUsage.EnergyUsage, v)
	})
}

func SetRealtimeCpuUsage(v *float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.ResourceUsage == nil {
			cd.RealtimeSystemData.ResourceUsage = &ClientResourceUsageData{}
		}
		return updateOptional(&cd.RealtimeSystemData.ResourceUsage.CpuUsage, v)
	})
}

func SetRealtimeMemUsage(v *float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.ResourceUsage == nil {
			cd.RealtimeSystemData.ResourceUsage = &ClientResourceUsageData{}
		}
		return updateOptional(&cd.RealtimeSystemData.ResourceUsage.MemUsage, v)
	})
}

func SetRealtimeNetworkUsage(v *float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.ResourceUsage == nil {
			cd.RealtimeSystemData.ResourceUsage = &ClientResourceUsageData{}
		}
		return updateOptional(&cd.RealtimeSystemData.ResourceUsage.NetworkUsage, v)
	})
}
