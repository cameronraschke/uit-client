package config

// Path-copy setters for Chassis port maps. These allocate new maps only when the
// value actually changes, and perform copy-on-write updates of the containing
// structs to preserve snapshot immutability.

import "maps"

// SetUSB1PortCount sets/updates the count for a labeled USB1 port.
func SetUSB1PortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		// Deep-copy hardware to avoid mutating shared snapshots
		hw := cloneClientHardwareData(cd.Hardware)
		if hw.Chassis == nil {
			hw.Chassis = &ChassisHardwareData{}
		}

		old := hw.Chassis.USB1Ports
		oldVal, exists := old[label]
		if exists && oldVal == count {
			return false
		}

		newMap := make(map[string]int64, len(old)+1)
		maps.Copy(newMap, old)
		newMap[label] = count

		hw.Chassis.USB1Ports = newMap
		cd.Hardware = hw
		return true
	})
}

// SetUSB2PortCount sets/updates the count for a labeled USB2 port.
func SetUSB2PortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := cloneClientHardwareData(cd.Hardware)
		if hw.Chassis == nil {
			hw.Chassis = &ChassisHardwareData{}
		}

		old := hw.Chassis.USB2Ports
		oldVal, exists := old[label]
		if exists && oldVal == count {
			return false
		}

		newMap := make(map[string]int64, len(old)+1)
		maps.Copy(newMap, old)
		newMap[label] = count

		hw.Chassis.USB2Ports = newMap
		cd.Hardware = hw
		return true
	})
}

// SetUSB3PortCount sets/updates the count for a labeled USB3 port.
func SetUSB3PortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := cloneClientHardwareData(cd.Hardware)
		if hw.Chassis == nil {
			hw.Chassis = &ChassisHardwareData{}
		}

		old := hw.Chassis.USB3Ports
		oldVal, exists := old[label]
		if exists && oldVal == count {
			return false
		}

		newMap := make(map[string]int64, len(old)+1)
		maps.Copy(newMap, old)
		newMap[label] = count

		hw.Chassis.USB3Ports = newMap
		cd.Hardware = hw
		return true
	})
}

// SetSATAPortCount sets/updates the count for a labeled SATA port.
func SetSATAPortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := cloneClientHardwareData(cd.Hardware)
		if hw.Chassis == nil {
			hw.Chassis = &ChassisHardwareData{}
		}

		old := hw.Chassis.SATAPorts
		oldVal, exists := old[label]
		if exists && oldVal == count {
			return false
		}

		newMap := make(map[string]int64, len(old)+1)
		maps.Copy(newMap, old)
		newMap[label] = count

		hw.Chassis.SATAPorts = newMap
		cd.Hardware = hw
		return true
	})
}

// SetInternalFanRPM sets/updates the RPM value for a labeled internal fan.
func SetInternalFanRPM(label string, rpm float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := cloneClientHardwareData(cd.Hardware)
		if hw.Chassis == nil {
			hw.Chassis = &ChassisHardwareData{}
		}

		old := hw.Chassis.InternalFans
		oldVal, exists := old[label]
		if exists && oldVal == rpm {
			return false
		}

		newMap := make(map[string]float64, len(old)+1)
		maps.Copy(newMap, old)
		newMap[label] = rpm

		hw.Chassis.InternalFans = newMap
		cd.Hardware = hw
		return true
	})
}

// SetAudioPortCount sets/updates the count for a labeled audio port.
func SetAudioPortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := cloneClientHardwareData(cd.Hardware)
		if hw.Chassis == nil {
			hw.Chassis = &ChassisHardwareData{}
		}

		old := hw.Chassis.AudioPorts
		oldVal, exists := old[label]
		if exists && oldVal == count {
			return false
		}

		newMap := make(map[string]int64, len(old)+1)
		maps.Copy(newMap, old)
		newMap[label] = count

		hw.Chassis.AudioPorts = newMap
		cd.Hardware = hw
		return true
	})
}
