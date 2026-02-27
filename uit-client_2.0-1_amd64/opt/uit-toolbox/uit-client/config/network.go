//go:build linux && amd64

package config

import (
	"net/netip"
	"uitclient/types"
)

func ipSliceEqual(a, b []netip.Addr) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalNetworkHardwareData(a, b types.NetworkHardwareData) bool {
	if (a.MACAddress == nil) != (b.MACAddress == nil) {
		return false
	}
	if a.MACAddress != nil && *a.MACAddress != *b.MACAddress {
		return false
	}

	if (a.Type == nil) != (b.Type == nil) {
		return false
	}
	if a.Type != nil && *a.Type != *b.Type {
		return false
	}

	if (a.Wired == nil) != (b.Wired == nil) {
		return false
	}
	if a.Wired != nil && *a.Wired != *b.Wired {
		return false
	}

	if (a.Wireless == nil) != (b.Wireless == nil) {
		return false
	}
	if a.Wireless != nil && *a.Wireless != *b.Wireless {
		return false
	}

	if (a.Model == nil) != (b.Model == nil) {
		return false
	}
	if a.Model != nil && *a.Model != *b.Model {
		return false
	}

	if (a.NetworkLinkUp == nil) != (b.NetworkLinkUp == nil) {
		return false
	}
	if a.NetworkLinkUp != nil && *a.NetworkLinkUp != *b.NetworkLinkUp {
		return false
	}

	// Compare IPAddress slice with nil==empty normalization
	if !ipSliceEqual(a.IPAddress, b.IPAddress) {
		return false
	}

	if (a.Netmask == nil) != (b.Netmask == nil) {
		return false
	}
	if a.Netmask != nil && *a.Netmask != *b.Netmask {
		return false
	}

	return true
}

func UpdateNetworkInterface(ifData map[string]types.NetworkHardwareData) {
	if len(ifData) == 0 {
		return
	}
	UpdateUniqueClientData(func(cd *types.ClientData) bool {
		if cd.Hardware == nil {
			cd.Hardware = &types.ClientHardwareData{}
		}
		if cd.Hardware.Network == nil {
			cd.Hardware.Network = make(map[string]types.NetworkHardwareData)
		}
		updated := false
		for ifName, newData := range ifData {
			oldData, exists := cd.Hardware.Network[ifName]
			if exists && equalNetworkHardwareData(oldData, newData) {
				continue
			}
			cd.Hardware.Network[ifName] = newData
			updated = true
		}
		return updated
	})
}
