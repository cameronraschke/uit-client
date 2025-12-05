package config

import "time"

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
