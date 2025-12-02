package config

import "maps"

func copyMap[K comparable, V any](oldMap map[K]V) map[K]V {
	if oldMap == nil {
		return nil
	}
	newMap := make(map[K]V, len(oldMap))
	maps.Copy(newMap, oldMap)
	return newMap
}

func cloneClientHardwareData(hardwareData *ClientHardwareData) *ClientHardwareData {
	if hardwareData == nil {
		return nil
	}
	copyHardwareData := *hardwareData

	if hardwareData.CPU != nil {
		cpuCopy := *hardwareData.CPU
		cpuCopy.ThermalProbeWorking = copyMap(hardwareData.CPU.ThermalProbeWorking)
		copyHardwareData.CPU = &cpuCopy
	}
	if hardwareData.Motherboard != nil {
		motherboardCopy := *hardwareData.Motherboard
		motherboardCopy.PCIELanes = copyMap(hardwareData.Motherboard.PCIELanes)
		motherboardCopy.M2Slots = copyMap(hardwareData.Motherboard.M2Slots)
		motherboardCopy.ThermalProbeWorking = copyMap(hardwareData.Motherboard.ThermalProbeWorking)
		copyHardwareData.Motherboard = &motherboardCopy
	}
	if hardwareData.Memory != nil {
		copyHardwareData.Memory = copyMap(hardwareData.Memory)
	}
	if hardwareData.Network != nil {
		copyHardwareData.Network = copyMap(hardwareData.Network)
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
		copyChassisData.UBS1Ports = copyMap(hardwareData.Chassis.UBS1Ports)
		copyChassisData.USB2Ports = copyMap(hardwareData.Chassis.USB2Ports)
		copyChassisData.USB3Ports = copyMap(hardwareData.Chassis.USB3Ports)
		copyChassisData.SATAPorts = copyMap(hardwareData.Chassis.SATAPorts)
		copyChassisData.InternalFans = copyMap(hardwareData.Chassis.InternalFans)
		copyChassisData.AudioPorts = copyMap(hardwareData.Chassis.AudioPorts)
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
