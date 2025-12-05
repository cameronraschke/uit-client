//go:build linux && amd64

package client

func GetBluetoothData() bool {
	// /sys/class/bluetooth/hci0
	return false
}

func GetWirelessData() bool {
	// /sys/class/ieee80211/phy0
	return false
}
