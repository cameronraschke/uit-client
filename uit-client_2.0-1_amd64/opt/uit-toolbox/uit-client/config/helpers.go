//go:build linux && amd64

package config

import (
	"maps"
	"net/netip"
)

func updateOptional[T comparable](dst **T, newVal *T) bool {
	if newVal == nil {
		if *dst == nil {
			return false
		}
		*dst = nil // Clear the destination pointer
		return true
	}
	val := *newVal // Dereference newVal to get the value
	if *dst != nil && **dst == val {
		return false // If values are the same, no update needed
	}
	// Allocate a fresh copy so callers can't mutate our internal state
	valCopy := val
	*dst = &valCopy // Store pointer to the new copy
	return true
}

func deepCopyMap[K comparable, V any](oldMap map[K]V) map[K]V {
	if oldMap == nil {
		return nil
	}
	newMap := make(map[K]V, len(oldMap))
	maps.Copy(newMap, oldMap)
	return newMap
}

// Specialized map copy helpers
func deepCopyInt64Map(srcMap map[string]int64) map[string]int64 {
	if srcMap == nil {
		return nil
	}
	dstMap := make(map[string]int64, len(srcMap))
	maps.Copy(dstMap, srcMap)
	return dstMap
}

func deepCopyFloat64Map(srcMap map[string]float64) map[string]float64 {
	if srcMap == nil {
		return nil
	}
	dstMap := make(map[string]float64, len(srcMap))
	maps.Copy(dstMap, srcMap)
	return dstMap
}

// nil => nil, true/false => new pointer with the same value.
func deepCopyBoolPtrMap(srcMap map[string]*bool) map[string]*bool {
	if srcMap == nil {
		return nil
	}
	dstMap := make(map[string]*bool, len(srcMap))
	for k, v := range srcMap {
		if v == nil {
			dstMap[k] = nil
		} else {
			val := *v
			vPtr := new(bool)
			*vPtr = val
			dstMap[k] = vPtr
		}
	}
	return dstMap
}

func deepCopyNetipAddrs(srcSlice []netip.Addr) []netip.Addr {
	if srcSlice == nil {
		return nil
	}
	dstSlice := make([]netip.Addr, len(srcSlice))
	copy(dstSlice, srcSlice)
	return dstSlice
}

func deepCopyNetworkHardwareMap(srcMap map[string]NetworkHardwareData) map[string]NetworkHardwareData {
	if srcMap == nil {
		return nil
	}
	dstMap := make(map[string]NetworkHardwareData, len(srcMap))
	for k, v := range srcMap {
		entry := v
		entry.IPAddress = deepCopyNetipAddrs(v.IPAddress)
		dstMap[k] = entry
	}
	return dstMap
}

func deepCopyClientHardwareData(hardwareData *ClientHardwareData) *ClientHardwareData {
	if hardwareData == nil {
		return nil
	}
	copyHardwareData := *hardwareData // Shallow copy of the struct

	if hardwareData.CPU != nil {
		cpuCopy := *hardwareData.CPU
		cpuCopy.ThermalProbeWorking = deepCopyBoolPtrMap(hardwareData.CPU.ThermalProbeWorking)
		copyHardwareData.CPU = &cpuCopy
	}
	if hardwareData.Motherboard != nil {
		motherboardCopy := *hardwareData.Motherboard
		motherboardCopy.PCIELanes = deepCopyInt64Map(hardwareData.Motherboard.PCIELanes)
		motherboardCopy.M2Slots = deepCopyInt64Map(hardwareData.Motherboard.M2Slots)
		motherboardCopy.ThermalProbeWorking = deepCopyBoolPtrMap(hardwareData.Motherboard.ThermalProbeWorking)
		copyHardwareData.Motherboard = &motherboardCopy
	}
	if hardwareData.Memory != nil {
		copyHardwareData.Memory = deepCopyMap(hardwareData.Memory)
	}
	if hardwareData.Network != nil {
		newNet := deepCopyNetworkHardwareMap(hardwareData.Network)
		copyHardwareData.Network = newNet
	}
	if hardwareData.Graphics != nil {
		copyGraphics := *hardwareData.Graphics
		copyHardwareData.Graphics = &copyGraphics
	}
	if hardwareData.Disks != nil {
		copyHardwareData.Disks = deepCopyMap(hardwareData.Disks)
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
		copyChassisData.USB1Ports = deepCopyInt64Map(hardwareData.Chassis.USB1Ports)
		copyChassisData.USB2Ports = deepCopyInt64Map(hardwareData.Chassis.USB2Ports)
		copyChassisData.USB3Ports = deepCopyInt64Map(hardwareData.Chassis.USB3Ports)
		copyChassisData.SATAPorts = deepCopyInt64Map(hardwareData.Chassis.SATAPorts)
		copyChassisData.InternalFans = deepCopyFloat64Map(hardwareData.Chassis.InternalFans)
		copyChassisData.AudioPorts = deepCopyInt64Map(hardwareData.Chassis.AudioPorts)
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

func deepCopyClientSoftwareData(softwareData *ClientSoftwareData) *ClientSoftwareData {
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

func deepCopyRealtimeSystemData(realtimeData *RealtimeSystemData) *RealtimeSystemData {
	if realtimeData == nil {
		return nil
	}
	copyRealtimeData := *realtimeData

	if realtimeData.Hardware != nil {
		copyRealtimeData.Hardware = deepCopyClientHardwareData(realtimeData.Hardware)
	}
	if realtimeData.Software != nil {
		copyRealtimeData.Software = deepCopyClientSoftwareData(realtimeData.Software)
	}
	if realtimeData.ResourceUsage != nil {
		resourceUsageCopy := *realtimeData.ResourceUsage
		copyRealtimeData.ResourceUsage = &resourceUsageCopy
	}
	return &copyRealtimeData
}

func deepCopyJobData(jobData *JobData) *JobData {
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

func copyClientData(clientData *ClientData) *ClientData {
	if clientData == nil {
		return &ClientData{}
	}
	// Copies all fields, but performs deep copy on pointer and map fields.
	copyClientData := *clientData
	if clientData.Hardware != nil {
		copyClientData.Hardware = deepCopyClientHardwareData(clientData.Hardware)
	}
	if clientData.Software != nil {
		copyClientData.Software = deepCopyClientSoftwareData(clientData.Software)
	}
	if clientData.RealtimeSystemData != nil {
		copyClientData.RealtimeSystemData = deepCopyRealtimeSystemData(clientData.RealtimeSystemData)
	}
	if clientData.JobData != nil {
		copyClientData.JobData = deepCopyJobData(clientData.JobData)
	}
	return &copyClientData
}
