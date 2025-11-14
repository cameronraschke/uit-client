package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"uitclient/block"

	"golang.org/x/sys/cpu"
	"golang.org/x/sys/unix"
)

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

	blockDevices, err := block.ListBlockDevices("/dev")
	if err != nil {
		fmt.Printf("Error listing block devices: %v\n", err)
	}
	if blockDevices == nil {
		fmt.Printf("Block device list is nil\n")
	}
	if len(blockDevices) <= 0 {
		fmt.Printf("No block devices found\n")
	}

	var blockDeviceSelector = make(map[int64]string)
	fmt.Printf("Block Devices:\n")
	for i, device := range blockDevices {
		if device == nil {
			fmt.Printf("Block device at index %d is nil\n", i)
			continue
		}
		if device.Major == 0 {
			fmt.Printf("[%d] Name: %s, Path: %s, Device Type: %s\n",
				i, device.Name, device.Path, device.BlockDeviceType)
			blockDeviceSelector[int64(i)] = device.Path
		}
	}

	reader := bufio.NewReader(os.Stdin)
	inputtedDeviceIndex, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}
	chosenDevice, err := strconv.ParseInt(inputtedDeviceIndex, 10, 64)
	if err != nil {
		fmt.Printf("Error parsing input to integer: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chosen device name: %s, index: %d\n", blockDeviceSelector[chosenDevice], chosenDevice)
	// mountDir := unix.MountDir("source", "target", "fstype", 0, "data")

}
