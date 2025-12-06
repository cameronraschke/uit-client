//go:build linux && amd64

package cli

import (
	"fmt"
	"strconv"
	"uitclient/client"
)

func (cli *CLI) SelectBlockDevices() (string, int, error) {
	blockDevices, err := client.ListBlockDevices("/dev")
	if err != nil {
		cli.Echo("Error listing block devices: %v", err)
		return "", 0, fmt.Errorf("Error listing block devices: %v", err)
	}
	if blockDevices == nil {
		cli.Echo("Block device list is nil")
		return "", 0, fmt.Errorf("Block device list is nil")
	}
	if len(blockDevices) <= 0 {
		cli.Echo("No block devices found")
		return "", 0, fmt.Errorf("No block devices found")
	}

	var blockDeviceSelector = make(map[int]string)
	printIndex := 1
	for _, device := range blockDevices {
		if device == nil {
			cli.Echo("Block device entry is nil, skipping")
			continue
		}
		if device.LinuxMinorNumber != nil && *device.LinuxMinorNumber == 0 {
			if device.LinuxAlias == nil || *device.LinuxAlias == "" {
				cli.Echo("Block device has no alias, skipping")
				continue
			}
			if device.LinuxDevicePath == nil || *device.LinuxDevicePath == "" {
				cli.Echo("Block device has no device path, skipping")
				continue
			}
			if device.InterfaceType == nil || *device.InterfaceType == "" {
				cli.Echo("Block device has no interface type, skipping")
				continue
			}
			if device.CapacityMiB != nil && *device.CapacityMiB <= 0 {
				cli.Echo("Block device has zero or negative capacity, skipping")
				continue
			}
			cli.Echo("[%d] Name: %s, Path: %s, Device Type: %s, Capacity: %.2fGiB",
				printIndex, *device.LinuxAlias, *device.LinuxDevicePath, *device.InterfaceType, *device.CapacityMiB/1024)
			blockDeviceSelector[printIndex] = *device.LinuxDevicePath
			printIndex++
		}
	}

	if len(blockDeviceSelector) == 0 {
		cli.Echo("No suitable block devices found for selection")
		return "", 0, fmt.Errorf("No suitable block devices found for selection")
	}
	cli.Echo("Total block devices found: %d", len(blockDeviceSelector))
	inputtedDeviceIndex, err := cli.Read("Select a block device to use: ")
	if err != nil {
		cli.Echo("Error reading input: %v", err)
		return "", 0, fmt.Errorf("Error reading input: %v", err)
	}
	if inputtedDeviceIndex == "" {
		cli.Echo("No selection entered")
		return "", 0, fmt.Errorf("No selection entered")
	}
	var chosenDevice = -1
	chosenDevice, err = strconv.Atoi(inputtedDeviceIndex)
	if err != nil {
		cli.Echo("Error parsing input to integer: %v", err)
		return "", 0, fmt.Errorf("Error parsing input to integer: %v", err)
	}
	if chosenDevice < 1 {
		cli.Echo("Invalid device selection: %d", chosenDevice)
		return "", 0, fmt.Errorf("Invalid device selection: %d", chosenDevice)
	}
	path, ok := blockDeviceSelector[chosenDevice]
	if !ok || path == "" {
		cli.Echo("Selection %d not in list", chosenDevice)
		return "", 0, fmt.Errorf("Selection %d not in list", chosenDevice)
	}
	return path, len(blockDevices), nil
}
