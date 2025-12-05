//go:build linux && amd64

package config

import "maps"

func SetUSB1PortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		// Deep-copy hardware to avoid mutating shared snapshots
		hw := deepCopyClientHardwareData(cd.Hardware)
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

func SetUSB2PortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := deepCopyClientHardwareData(cd.Hardware)
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

func SetUSB3PortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := deepCopyClientHardwareData(cd.Hardware)
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

func SetSATAPortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := deepCopyClientHardwareData(cd.Hardware)
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

func SetInternalFanRPM(label string, rpm float64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := deepCopyClientHardwareData(cd.Hardware)
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

func SetAudioPortCount(label string, count int64) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			return false
		}

		hw := deepCopyClientHardwareData(cd.Hardware)
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
