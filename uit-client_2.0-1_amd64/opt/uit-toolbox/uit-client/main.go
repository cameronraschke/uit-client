package main

import (
	"fmt"
	"os"
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
	} else {
		fmt.Printf("Block Devices:\n")
		for _, device := range blockDevices {
			fmt.Printf("Name: %s, Path: %s, Major: %d, Minor: %d\n",
				device.Name, device.Path, device.Major, device.Minor)
		}
	}
	// mountDir := unix.MountDir("source", "target", "fstype", 0, "data")

}
