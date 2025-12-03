//go:build linux
// +build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"uitclient/hardware"
)

func selectBlockDevices() (string, int, error) {
	blockDevices, err := hardware.ListBlockDevices("/dev")
	if err != nil {
		return "", 0, fmt.Errorf("Error listing block devices: %v\n", err)
	}
	if blockDevices == nil {
		return "", 0, fmt.Errorf("Block device list is nil\n")
	}
	if len(blockDevices) <= 0 {
		return "", 0, fmt.Errorf("No block devices found\n")
	}

	var blockDeviceSelector = make(map[int]string)
	printIndex := 1
	for _, device := range blockDevices {
		if device == nil {
			fmt.Printf("Block device entry is nil, skipping\n")
			continue
		}
		if device.LinuxMinorNumber != nil && *device.LinuxMinorNumber == 0 {
			if device.LinuxAlias == nil || *device.LinuxAlias == "" {
				fmt.Printf("Block device has no alias, skipping\n")
				continue
			}
			if device.LinuxDevicePath == nil || *device.LinuxDevicePath == "" {
				fmt.Printf("Block device has no device path, skipping\n")
				continue
			}
			if device.InterfaceType == nil || *device.InterfaceType == "" {
				fmt.Printf("Block device has no interface type, skipping\n")
				continue
			}
			if device.CapacityMiB != nil && *device.CapacityMiB <= 0 {
				fmt.Printf("Block device has zero or negative capacity, skipping\n")
				continue
			}
			fmt.Printf("[%d] Name: %s, Path: %s, Device Type: %s, Capacity: %.2fGiB\n",
				printIndex, *device.LinuxAlias, *device.LinuxDevicePath, *device.InterfaceType, *device.CapacityMiB/1024)
			blockDeviceSelector[printIndex] = *device.LinuxDevicePath
			printIndex++
		}
	}

	if len(blockDeviceSelector) == 0 {
		return "", 0, fmt.Errorf("No suitable block devices found for selection\n")
	}
	fmt.Printf("Total block devices found: %d\n", len(blockDeviceSelector))
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect a block device to use: ")
	inputtedDeviceIndex, err := reader.ReadString('\n')
	if err != nil {
		return "", 0, fmt.Errorf("Error reading input: %v\n", err)
	}
	inputtedDeviceIndex = strings.TrimSpace(inputtedDeviceIndex)
	if inputtedDeviceIndex == "" {
		return "", 0, fmt.Errorf("No selection entered\n")
	}
	var chosenDevice = -1
	chosenDevice, err = strconv.Atoi(inputtedDeviceIndex)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing input to integer: %v\n", err)
	}
	if chosenDevice < 1 {
		return "", 0, fmt.Errorf("Invalid device selection: %d\n", chosenDevice)
	}
	path, ok := blockDeviceSelector[chosenDevice]
	if !ok || path == "" {
		return "", 0, fmt.Errorf("Selection %d not in list\n", chosenDevice)
	}
	return path, len(blockDevices), nil
}
