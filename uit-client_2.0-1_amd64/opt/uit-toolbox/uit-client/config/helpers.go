package config

import (
	"maps"
	"net/netip"
)

func updateOptional[T comparable](dst **T, incoming *T) bool {
	if incoming == nil {
		if *dst == nil {
			return false
		}
		*dst = nil // Clear the destination pointer
		return true
	}
	val := *incoming
	if *dst != nil && **dst == val {
		return false // If values are the same, no update needed
	}
	// Allocate a fresh copy so callers can't mutate our internal state
	valCopy := val
	*dst = &valCopy // Store pointer to the new copy
	return true
}

func copyMap[K comparable, V any](oldMap map[K]V) map[K]V {
	if oldMap == nil {
		return nil
	}
	newMap := make(map[K]V, len(oldMap))
	maps.Copy(newMap, oldMap)
	return newMap
}

// Specialized map copy helpers
func copyInt64Map(old map[string]int64) map[string]int64 {
	if old == nil {
		return nil
	}
	m := make(map[string]int64, len(old))
	maps.Copy(m, old)
	return m
}

func copyFloat64Map(old map[string]float64) map[string]float64 {
	if old == nil {
		return nil
	}
	m := make(map[string]float64, len(old))
	maps.Copy(m, old)
	return m
}

// copyBoolPtrMap performs a deep copy of map[string]*bool while preserving tri-state semantics:
// nil => nil, true/false => new pointer with the same value. This avoids pointer sharing across snapshots.
func copyBoolPtrMap(old map[string]*bool) map[string]*bool {
	if old == nil {
		return nil
	}
	m := make(map[string]*bool, len(old))
	for k, v := range old {
		if v == nil {
			m[k] = nil
		} else {
			val := *v
			vv := new(bool)
			*vv = val
			m[k] = vv
		}
	}
	return m
}

// copyNetipAddrs deep-copies a slice of netip.Addr to avoid sharing backing arrays.
func copyNetipAddrs(old []netip.Addr) []netip.Addr {
	if old == nil {
		return nil
	}
	out := make([]netip.Addr, len(old))
	copy(out, old)
	return out
}

func cloneClientHardwareData(hardwareData *ClientHardwareData) *ClientHardwareData {
	if hardwareData == nil {
		return nil
	}
	copyHardwareData := *hardwareData

	if hardwareData.CPU != nil {
		cpuCopy := *hardwareData.CPU
		cpuCopy.ThermalProbeWorking = copyBoolPtrMap(hardwareData.CPU.ThermalProbeWorking)
		copyHardwareData.CPU = &cpuCopy
	}
	if hardwareData.Motherboard != nil {
		motherboardCopy := *hardwareData.Motherboard
		motherboardCopy.PCIELanes = copyInt64Map(hardwareData.Motherboard.PCIELanes)
		motherboardCopy.M2Slots = copyInt64Map(hardwareData.Motherboard.M2Slots)
		motherboardCopy.ThermalProbeWorking = copyBoolPtrMap(hardwareData.Motherboard.ThermalProbeWorking)
		copyHardwareData.Motherboard = &motherboardCopy
	}
	if hardwareData.Memory != nil {
		copyHardwareData.Memory = copyMap(hardwareData.Memory)
	}
	if hardwareData.Network != nil {
		// Deep-copy network map and ensure IPAddress slice is copied per entry
		newNet := make(map[string]NetworkHardwareData, len(hardwareData.Network))
		for k, v := range hardwareData.Network {
			entry := v
			entry.IPAddress = copyNetipAddrs(v.IPAddress)
			newNet[k] = entry
		}
		copyHardwareData.Network = newNet
	}
	if hardwareData.Graphics != nil {
		copyGraphics := *hardwareData.Graphics
		copyHardwareData.Graphics = &copyGraphics
	}
	if hardwareData.Disks != nil {
		copyHardwareData.Disks = copyMap(hardwareData.Disks)
	}
	if hardwareData.Battery != nil {
		copyBattery := *hardwareData.Battery
		copyHardwareData.Battery = &copyBattery
	}
	if hardwareData.Wireless != nil {
		copyWireless := *hardwareData.Wireless
		copyHardwareData.Wireless = &copyWireless
	}
	if hardwareData.Chassis != nil {
		copyChassisData := *hardwareData.Chassis
		copyChassisData.USB1Ports = copyInt64Map(hardwareData.Chassis.USB1Ports)
		copyChassisData.USB2Ports = copyInt64Map(hardwareData.Chassis.USB2Ports)
		copyChassisData.USB3Ports = copyInt64Map(hardwareData.Chassis.USB3Ports)
		copyChassisData.SATAPorts = copyInt64Map(hardwareData.Chassis.SATAPorts)
		copyChassisData.InternalFans = copyFloat64Map(hardwareData.Chassis.InternalFans)
		copyChassisData.AudioPorts = copyInt64Map(hardwareData.Chassis.AudioPorts)
		copyHardwareData.Chassis = &copyChassisData
	}
	if hardwareData.PowerSupply != nil {
		copyPowerSupply := *hardwareData.PowerSupply
		copyHardwareData.PowerSupply = &copyPowerSupply
	}
	if hardwareData.TPM != nil {
		copyTPM := *hardwareData.TPM
		copyHardwareData.TPM = &copyTPM
	}
	return &copyHardwareData
}

func cloneClientSoftwareData(softwareData *ClientSoftwareData) *ClientSoftwareData {
	if softwareData == nil {
		return nil
	}
	copySoftwareData := *softwareData

	if softwareData.Motherboard != nil {
		mbSoftwareDataCopy := *softwareData.Motherboard
		copySoftwareData.Motherboard = &mbSoftwareDataCopy
	}
	return &copySoftwareData
}

func cloneRealtimeSystemData(realtimeData *RealtimeSystemData) *RealtimeSystemData {
	if realtimeData == nil {
		return nil
	}
	copyRealtimeData := *realtimeData

	if realtimeData.Hardware != nil {
		copyRealtimeData.Hardware = cloneClientHardwareData(realtimeData.Hardware)
	}
	if realtimeData.Software != nil {
		copyRealtimeData.Software = cloneClientSoftwareData(realtimeData.Software)
	}
	if realtimeData.ResourceUsage != nil {
		resourceUsageCopy := *realtimeData.ResourceUsage
		copyRealtimeData.ResourceUsage = &resourceUsageCopy
	}
	return &copyRealtimeData
}

func cloneJobData(jobData *JobData) *JobData {
	if jobData == nil {
		return nil
	}
	copyJobData := *jobData

	if jobData.AvgResourceUsage != nil {
		copyResourceUsage := *jobData.AvgResourceUsage
		copyJobData.AvgResourceUsage = &copyResourceUsage
	}
	if jobData.Realtime != nil {
		copyRealtimeJobData := *jobData.Realtime
		copyJobData.Realtime = &copyRealtimeJobData
	}
	return &copyJobData
}

func cloneClientData(clientData *ClientData) *ClientData {
	if clientData == nil {
		return &ClientData{}
	}
	// Copies all fields, but performs deep copy on pointer and map fields.
	copyClientData := *clientData
	if clientData.Hardware != nil {
		copyClientData.Hardware = cloneClientHardwareData(clientData.Hardware)
	}
	if clientData.Software != nil {
		copyClientData.Software = cloneClientSoftwareData(clientData.Software)
	}
	if clientData.RealtimeSystemData != nil {
		copyClientData.RealtimeSystemData = cloneRealtimeSystemData(clientData.RealtimeSystemData)
	}
	if clientData.JobData != nil {
		copyClientData.JobData = cloneJobData(clientData.JobData)
	}
	return &copyClientData
}
