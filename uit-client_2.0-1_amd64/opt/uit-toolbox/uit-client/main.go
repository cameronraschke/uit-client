//go:build linux
// +build linux

package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"uitclient/config"
	"uitclient/hardware"
	"uitclient/webclient"

	"golang.org/x/sys/cpu"
	"golang.org/x/sys/unix"
)

var dbConn *sql.DB

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
		if device.Minor == 0 {
			fmt.Printf("[%d] Name: %s, Path: %s, Device Type: %s, Capacity: %.2fGiB\n",
				printIndex, device.Name, device.Path, device.DiskType, device.CapacityMiB/1024)
			blockDeviceSelector[printIndex] = device.Path
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

func main() {
	euid := unix.Geteuid()
	if euid > 1000 {
		fmt.Printf("Please run as root, current EUID: %d", euid)
		os.Exit(1)
	}
	var err error
	pid := unix.Getpid()
	parentPid := unix.Getppid()

	fmt.Printf("EUID: %d, PID: %d, Parent PID: %d\n", euid, pid, parentPid)

	clientConfigJson, err := webclient.GetClientConfig()
	if err != nil {
		fmt.Printf("Error getting client configuration: %v\n", err)
		os.Exit(1)
	}
	tmpConfig := &config.ClientConfig{}
	if err = json.Unmarshal(clientConfigJson, tmpConfig); err != nil {
		fmt.Printf("Error unmarshaling client configuration JSON: %v\n", err)
		os.Exit(1)
	}
	err = config.InitializeClientConfig(tmpConfig)
	if err != nil {
		fmt.Printf("Error initializing client configuration: %v\n", err)
		os.Exit(1)
	}

	clientConfig := config.GetClientConfig()
	if clientConfig == nil {
		fmt.Printf("Client configuration is nil after initialization\n")
		os.Exit(1)
	}

	fmt.Printf("Client configuration loaded: %+v\n", clientConfig)

	systemSerial, err := hardware.GetSystemSerial()
	if err != nil {
		fmt.Printf("Error getting system serial: %v\n", err)
		os.Exit(1)
	}
	if systemSerial == "" {
		fmt.Printf("System serial number is empty\n")
		os.Exit(1)
	}

	tagnumber, err := webclient.SerialLookup(systemSerial)
	if err != nil {
		fmt.Printf("Error looking up serial number: %v\n", err)
		os.Exit(1)
	}
	if tagnumber < 1 {
		fmt.Printf("Invalid tagnumber retrieved: %d\n", tagnumber)
		os.Exit(1)
	}

	fmt.Printf("Tagnumber: %d, System Serial: %s\n", tagnumber, systemSerial)

	hasFP := cpu.X86.HasAES || cpu.ARM64.HasSHA1 || cpu.ARM64.HasSHA2 || cpu.ARM64.HasSHA3 || cpu.ARM64.HasCRC32
	if hasFP {
		fmt.Printf("CPU has encryption acceleration\n")
	}

	var statfs unix.Statfs_t
	if err := unix.Statfs("/", &statfs); err != nil {
		fmt.Printf("Statfs error: %v\n", err)
	}

	fmt.Printf("Filesystem type: %x\n", statfs.Type)

	devicePath, totalDevices, err := selectBlockDevices()
	if err != nil {
		fmt.Printf("Error selecting block device: %v\n", err)
		os.Exit(1)
	}
	if totalDevices <= 0 {
		fmt.Printf("No block devices found on system\n")
		os.Exit(1)
	}

	fmt.Printf("Selected block device path: %s\n", devicePath)
}
