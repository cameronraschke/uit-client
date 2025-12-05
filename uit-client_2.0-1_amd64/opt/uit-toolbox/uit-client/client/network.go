//go:build linux && amd64

package client

func GetInterfaceData() map[string]map[string]string {
	// Read network interface data from /sys/class/net/
	return map[string]map[string]string{}
}
