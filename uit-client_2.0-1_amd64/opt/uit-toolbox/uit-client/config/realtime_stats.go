//go:build linux && amd64

package config

import (
	"time"
	"uitclient/types"
)

func UpdateRealtimeJobQueue(mutate func(*types.JobQueueData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		mutate(cd.JobData.Realtime)
		return true
	})
}

func SetRealtimeJobName(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobName, v)
	})
}

func SetRealtimeJobNameFormatted(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobNameFormatted, v)
	})
}

func SetRealtimeJobQueued(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueued, v)
	})
}

func SetRealtimeJobRequiresQueue(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobRequiresQueue, v)
	})
}

func SetRealtimeJobQueuePosition(v *int) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueuePosition, v)
	})
}

func SetRealtimeJobQueuedOverride(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobQueuedOverride, v)
	})
}

func SetRealtimeJobActive(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobActive, v)
	})
}

func SetRealtimeJobProgress(v *float64) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobProgress, v)
	})
}

func SetRealtimeJobDuration(d time.Duration) {
	if d.Seconds() < 0 {
		return
	}
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		if cd.JobData.Realtime.JobDuration == d {
			return false
		}
		cd.JobData.Realtime.JobDuration = d
		return true
	})
}

func SetRealtimeJobStatusMessage(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.JobData == nil {
			cd.JobData = &types.JobData{}
		}
		if cd.JobData.Realtime == nil {
			cd.JobData.Realtime = &types.JobQueueData{}
		}
		return updateOptional(&cd.JobData.Realtime.JobStatusMessage, v)
	})
}

// UpdateRealtimeSystemData mutates the realtime system subtree using copy-on-write.
func UpdateRealtimeSystemData(mutate func(*types.RealtimeSystemData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		mutate(cd.RealtimeSystemData)
		return true
	})
}

// --- RealtimeSystemData setters ---

func SetLastHeardTimestamp(t *time.Time) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.LastHeardTimestamp, t)
	})
}

func SetSystemUptime(uptime time.Duration) {
	if uptime.Seconds() < 0 {
		return
	}
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.SystemUptime == uptime {
			return false
		}
		cd.RealtimeSystemData.SystemUptime = uptime
		return true
	})
}

func SetAppUptime(appUptime *time.Duration) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.AppUptime, appUptime)
	})
}

func SetKernelUpdated(kernelUpdated *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		return updateOptional(&cd.RealtimeSystemData.KernelUpdated, kernelUpdated)
	})
}

// --- Realtime ResourceUsage setters ---

func SetRealtimeCpuUsage(cpuUsage *types.CPUUsage) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.CPU == nil {
			cd.RealtimeSystemData.CPU = &types.CPUUsage{}
		}
		return updateOptional(&cd.RealtimeSystemData.CPU, cpuUsage)
	})
}

func SetRealtimeMemUsage(memUsage *types.MemoryUsage) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.Memory == nil {
			cd.RealtimeSystemData.Memory = &types.MemoryUsage{}
		}
		return updateOptional(&cd.RealtimeSystemData.Memory, memUsage)
	})
}

func SetRealtimeNetworkUsage(netUsage *types.NetworkUsage) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.Network == nil {
			cd.RealtimeSystemData.Network = &types.NetworkUsage{}
		}
		return updateOptional(&cd.RealtimeSystemData.Network, netUsage)
	})
}

func SetRealtimeEnergyUsage(energyUsage *types.EnergyUsage) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.RealtimeSystemData == nil {
			cd.RealtimeSystemData = &types.RealtimeSystemData{}
		}
		if cd.RealtimeSystemData.Energy == nil {
			cd.RealtimeSystemData.Energy = &types.EnergyUsage{}
		}
		return updateOptional(&cd.RealtimeSystemData.Energy, energyUsage)
	})
}
