//go:build linux && amd64

package config

import (
	"slices"
	"uitclient/types"
)

func UpdateHardwareData(mutate func(*types.ClientHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		mutate(cd.Hardware) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateCPUHardware(mutate func(*types.CPUHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.CPU == nil {
			cd.Hardware.CPU = &types.CPUHardwareData{}
		}
		mutate(cd.Hardware.CPU) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateMotherboardHardware(mutate func(*types.MotherboardHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Motherboard == nil {
			cd.Hardware.Motherboard = &types.MotherboardHardwareData{}
		}
		mutate(cd.Hardware.Motherboard) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateMemoryHardware(ramSerial string, mutate func(*types.MemoryHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Memory == nil {
			cd.Hardware.Memory = make(map[string]types.MemoryHardwareData)
		}
		originalCopy, existed := cd.Hardware.Memory[ramSerial]
		newCopy := originalCopy
		mutate(&newCopy)
		if existed && newCopy == originalCopy {
			return false
		}
		cd.Hardware.Memory[ramSerial] = newCopy
		return true
	})
}

func UpdateNetworkHardware(ifName string, mutate func(*types.NetworkHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Network == nil {
			cd.Hardware.Network = make(map[string]types.NetworkHardwareData)
		}
		originalCopy, existed := cd.Hardware.Network[ifName]
		newCopy := originalCopy
		mutate(&newCopy)
		if existed && slices.Equal(newCopy.IPAddress, originalCopy.IPAddress) {
			return false
		}
		cd.Hardware.Network[ifName] = newCopy
		return true
	})
}

func UpdateGraphicsHardware(mutate func(*types.GraphicsHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Graphics == nil {
			cd.Hardware.Graphics = &types.GraphicsHardwareData{}
		}
		mutate(cd.Hardware.Graphics) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateDiskHardware(diskName string, mutate func(*types.DiskHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Disks == nil {
			cd.Hardware.Disks = make(map[string]types.DiskHardwareData)
		}
		originalCopy, existed := cd.Hardware.Disks[diskName]
		newCopy := originalCopy
		mutate(&newCopy)
		if existed && newCopy == originalCopy {
			return false
		}
		cd.Hardware.Disks[diskName] = newCopy
		return true
	})
}

func UpdateBatteryHardware(mutate func(*types.BatteryHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Battery == nil {
			cd.Hardware.Battery = &types.BatteryHardwareData{}
		}
		mutate(cd.Hardware.Battery) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateWirelessHardware(mutate func(*types.WirelessHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Wireless == nil {
			cd.Hardware.Wireless = &types.WirelessHardwareData{}
		}
		mutate(cd.Hardware.Wireless) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateChassisHardware(mutate func(*types.ChassisHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Chassis == nil {
			cd.Hardware.Chassis = &types.ChassisHardwareData{}
		}
		mutate(cd.Hardware.Chassis) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdatePowerSupplyHardware(mutate func(*types.PowerSupplyHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.PowerSupply == nil {
			cd.Hardware.PowerSupply = &types.PowerSupplyHardwareData{}
		}
		mutate(cd.Hardware.PowerSupply) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateTPMHardware(mutate func(*types.TPMHardwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.TPM == nil {
			cd.Hardware.TPM = &types.TPMHardwareData{}
		}
		mutate(cd.Hardware.TPM) // mutate in place, already isolated by copy-on-write
		return true
	})
}
