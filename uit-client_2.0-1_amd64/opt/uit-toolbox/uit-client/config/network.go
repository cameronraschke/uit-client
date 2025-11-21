package config

func UpdateMacAddress(ifData map[string]NetworkHardwareData) {
	if len(ifData) == 0 {
		return
	}
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &ClientHardwareData{}
		}
		if cd.Hardware.Network == nil {
			cd.Hardware.Network = make(map[string]NetworkHardwareData, len(ifData))
		}
		changed := false
		for k, v := range ifData {
			if entry, ok := cd.Hardware.Network[k]; !ok || entry != v {
				cd.Hardware.Network[k] = v
				changed = true
			}
		}
		return changed
	})
}
