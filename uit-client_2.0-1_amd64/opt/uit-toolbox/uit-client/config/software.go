package config

func UpdateSoftwareData(mutate func(*ClientSoftwareData)) {
	UpdateUniqueClientData(func(cd *ClientData) bool {
		if cd.Software == nil {
			cd.Software = &ClientSoftwareData{}
		}
		mutate(cd.Software) // mutate in place, already isolated by copy-on-write
		return true
	})
}
