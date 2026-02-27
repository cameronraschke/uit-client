//go:build linux && amd64

package config

import (
	"time"
	"uitclient/types"
)

// UpdateSoftwareData mutates the software subtree using copy-on-write.
func UpdateSoftwareData(mutate func(*types.ClientSoftwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		mutate(cd.Software)
		return true
	})
}

// --- types.ClientSoftwareData setters ---

func SetOSInstalled(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		return updateOptional(&cd.Software.OSInstalled, v)
	})
}

func SetOSName(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		return updateOptional(&cd.Software.OSName, v)
	})
}

func SetOSVersion(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		return updateOptional(&cd.Software.OSVersion, v)
	})
}

func SetOSInstalledTimestamp(t *time.Time) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		return updateOptional(&cd.Software.OSInstalledTimestamp, t)
	})
}

func SetImageName(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		return updateOptional(&cd.Software.ImageName, v)
	})
}

// UpdateMotherboardSoftwareData mutates the motherboard software subtree.
func UpdateMotherboardSoftwareData(mutate func(*types.MotherboardSoftwareData)) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		mutate(cd.Software.Motherboard)
		return true
	})
}

// --- types.MotherboardSoftwareData setters ---

func SetBIOSUpdated(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.BIOSUpdated, v)
	})
}

func SetBIOSVersion(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.BIOSVersion, v)
	})
}

func SetBIOSDate(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.BIOSDate, v)
	})
}

func SetBIOSFirmwareRevision(v *string) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.BIOSFirmwareRevision, v)
	})
}

func SetUEFIEnabled(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.UEFIEnabled, v)
	})
}

func SetSecureBootEnabled(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.SecureBootEnabled, v)
	})
}

func SetTPMEnabled(v *bool) {
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Software == nil {
			cd.Software = &types.ClientSoftwareData{}
		}
		if cd.Software.Motherboard == nil {
			cd.Software.Motherboard = &types.MotherboardSoftwareData{}
		}
		return updateOptional(&cd.Software.Motherboard.TPMEnabled, v)
	})
}
