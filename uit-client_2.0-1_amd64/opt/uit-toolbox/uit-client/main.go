package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"uitclient/hardware"

	"golang.org/x/sys/cpu"
	"golang.org/x/sys/unix"
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
	for i, device := range blockDevices {
		if device == nil {
			fmt.Printf("Block device at index %d is nil\n", i)
			continue
		}
		if device.Minor == 0 {
			fmt.Printf("[%d] Name: %s, Path: %s, Device Type: %s, Capacity: %.2fGiB\n",
				i, device.Name, device.Path, device.DiskType, device.CapacityMiB/1024)
			blockDeviceSelector[i] = device.Path
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
	inputtedDeviceIndex = inputtedDeviceIndex[:len(inputtedDeviceIndex)-1] // Remove newline character from read
	var chosenDevice = -1
	chosenDevice, err = strconv.Atoi(inputtedDeviceIndex)
	if err != nil {
		return "", 0, fmt.Errorf("Error parsing input to integer: %v\n", err)
	}
	if len(blockDeviceSelector[chosenDevice]) == 0 || blockDeviceSelector[chosenDevice] == "" || chosenDevice < 0 {
		return "", 0, fmt.Errorf("Invalid device selection: %d\n", chosenDevice)
	}
	return blockDeviceSelector[chosenDevice], len(blockDevices), nil
}

func main() {
	euid := unix.Geteuid()
	if euid > 1000 {
		fmt.Printf("Please run as root, current EUID: %d", euid)
		os.Exit(1)
	}
	pid := unix.Getpid()
	parentPid := unix.Getppid()

	fmt.Printf("EUID: %d, PID: %d, Parent PID: %d\n", euid, pid, parentPid)

	hasFP := cpu.X86.HasAES || cpu.ARM64.HasSHA1 || cpu.ARM64.HasSHA2 || cpu.ARM64.HasSHA3 || cpu.ARM64.HasCRC32
	if hasFP {
		fmt.Printf("CPU has encryption acceleration\n")
	}

	var statfs unix.Statfs_t
	if err := unix.Statfs("/", &statfs); err != nil {
		fmt.Printf("Statfs error: %v\n", err)
	}

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
