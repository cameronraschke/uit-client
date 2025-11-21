package config

func UpdateHardwareData(mutate func(*ClientHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		mutate(cd.Hardware) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateCPUHardware(mutate func(*CPUHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.CPU == nil {
			cd.Hardware.CPU = &CPUHardwareData{}
		}
		mutate(cd.Hardware.CPU) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateMotherboardHardware(mutate func(*MotherboardHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Motherboard == nil {
			cd.Hardware.Motherboard = &MotherboardHardwareData{}
		}
		mutate(cd.Hardware.Motherboard) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateMemoryHardware(ramSerial string, mutate func(*MemoryHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Memory == nil {
			cd.Hardware.Memory = make(map[string]MemoryHardwareData)
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

func UpdateNetworkHardware(macAddr string, mutate func(*NetworkHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Network == nil {
			cd.Hardware.Network = make(map[string]NetworkHardwareData)
		}
		originalCopy, existed := cd.Hardware.Network[macAddr]
		newCopy := originalCopy
		mutate(&newCopy)
		if existed && newCopy == originalCopy {
			return false
		}
		cd.Hardware.Network[macAddr] = newCopy
		return true
	})
}

func UpdateGraphicsHardware(mutate func(*GraphicsHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Graphics == nil {
			cd.Hardware.Graphics = &GraphicsHardwareData{}
		}
		mutate(cd.Hardware.Graphics) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateDiskHardware(diskName string, mutate func(*DiskHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Disks == nil {
			cd.Hardware.Disks = make(map[string]DiskHardwareData)
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

func UpdateBatteryHardware(mutate func(*BatteryHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Battery == nil {
			cd.Hardware.Battery = &BatteryHardwareData{}
		}
		mutate(cd.Hardware.Battery) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateWirelessHardware(mutate func(*WirelessHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Wireless == nil {
			cd.Hardware.Wireless = &WirelessHardwareData{}
		}
		mutate(cd.Hardware.Wireless) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateChassisHardware(mutate func(*ChassisHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Chassis == nil {
			cd.Hardware.Chassis = &ChassisHardwareData{}
		}
		mutate(cd.Hardware.Chassis) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdatePowerSupplyHardware(mutate func(*PowerSupplyHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.PowerSupply == nil {
			cd.Hardware.PowerSupply = &PowerSupplyHardwareData{}
		}
		mutate(cd.Hardware.PowerSupply) // mutate in place, already isolated by copy-on-write
		return true
	})
}

func UpdateTPMHardware(mutate func(*TPMHardwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.TPM == nil {
			cd.Hardware.TPM = &TPMHardwareData{}
		}
		mutate(cd.Hardware.TPM) // mutate in place, already isolated by copy-on-write
		return true
	})
}
