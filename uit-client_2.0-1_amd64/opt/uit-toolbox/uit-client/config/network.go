package config

import "reflect"

func equalNetworkHardwareData(a, b NetworkHardwareData) bool {
	return reflect.DeepEqual(a, b)
}

func UpdateMacAddress(ifData map[string]NetworkHardwareData) {
	if len(ifData) == 0 {
		return
	}
	UpdateUniqueClientData(func(cd *ClientData) bool {
		changed := false
		for k, v := range ifData {
			entry, ok := cd.Hardware.Network[k]
			if !ok || !equalNetworkHardwareData(entry, v) {
				cd.Hardware.Network[k] = v
				changed = true
			}
		}
		return changed
	})
}
