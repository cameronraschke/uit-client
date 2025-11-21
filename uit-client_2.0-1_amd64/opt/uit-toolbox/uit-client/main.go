//go:build linux
// +build linux

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"uitclient/config"
	"uitclient/hardware"
	"uitclient/webclient"

	"golang.org/x/sys/cpu"
	"golang.org/x/sys/unix"
)

const clearScreen = `\e[1;1H\e[2J`

// const clearScreen = "\033[H\033[2J"

func getClientData() {
	// System data
	systemSerial, err := hardware.GetSystemSerial()
	if err != nil {
		fmt.Printf("Error getting system serial: %v\n", err)
		os.Exit(1)
	}
	if systemSerial == "" {
		fmt.Printf("System serial number is empty\n")
		os.Exit(1)
	}
	config.SetSystemSerial(systemSerial)

	tagnumber, err := webclient.SerialLookup(systemSerial)
	if err != nil {
		fmt.Printf("Error looking up serial number: %v\n", err)
		os.Exit(1)
	}
	if tagnumber < 1 {
		fmt.Printf("Invalid tagnumber retrieved: %d\n", tagnumber)
		os.Exit(1)
	}
	config.SetTagnumber(tagnumber)

	systemUUID, err := hardware.GetSystemUUID()
	if err != nil {
		fmt.Printf("Error getting system UUID: %v\n", err)
		os.Exit(1)
	}
	config.SetSystemUUID(systemUUID)

	manufacturer := hardware.GetSystemManufacturer()
	config.SetManufacturer(manufacturer)

	model := hardware.GetSystemModel()
	config.SetModel(model)

	sku := hardware.GetSystemSKU()
	config.SetSKU(sku)

	// Network data
	connectedToHost := true // assumed, got client config from server
	config.SetConnectedToHost(&connectedToHost)

	networkInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting network interfaces: %v\n", err)
		os.Exit(1)
	}

	networkMap := make(map[string]config.NetworkHardwareData)
	for _, netIf := range networkInterfaces {
		ifName := netIf.Name
		if ifName == "" {
			fmt.Printf("Network interface has no name, skipping\n")
			continue
		}
		macAddress := netIf.HardwareAddr.String()
		if macAddress == "" {
			fmt.Printf("Interface %s has no MAC address, skipping\n", ifName)
			continue
		}
		ipAddresses, err := netIf.Addrs()
		if err != nil {
			fmt.Printf("Error getting IP addresses for interface %s: %v\n", ifName, err)
			continue
		}

		ipAddressSlice := []netip.Addr{}
		for _, addr := range ipAddresses {
			ipAddr, err := netip.ParseAddr(addr.String())
			if err != nil {
				fmt.Printf("Error parsing IP address %s for interface %s: %v\n", addr.String(), ifName, err)
				continue
			}
			ipAddressSlice = append(ipAddressSlice, ipAddr)
		}

		linkUp := (netIf.Flags & net.FlagUp) != 0
		networkMap[ifName] = config.NetworkHardwareData{
			MACAddress:    macAddress,
			NetworkLinkUp: &linkUp,
			IPAddress:     ipAddressSlice,
		}

	}

	// Job data
	jobUUID, err := config.CreateNewJobUUID()
	if err != nil {
		fmt.Printf("Error creating new job UUID: %v\n", err)
		os.Exit(1)
	}
	config.SetJobUUID(jobUUID)

	config.UpdateClientData(func(clientData *config.ClientData) {
		clientData.Serial = systemSerial
	})
}

func main() {
	recover()
	var err error
	// Initial startup, checks, and configuration loading

	// Clear the terminal screen
	fmt.Printf(clearScreen)
	fmt.Printf("Starting UIT Client...\n\n")

	// Check for root privileges & PIDs
	euid := unix.Geteuid()
	if euid > 1000 {
		fmt.Printf("Please run as root, current EUID: %d", euid)
		os.Exit(1)
	}
	// pid := unix.Getpid()
	// parentPid := unix.Getppid()
	// fmt.Printf("EUID: %d, PID: %d, Parent PID: %d\n", euid, pid, parentPid)

	// Fetch and initialize client configuration
	clientConfigJson, err := webclient.GetClientConfig()
	if err != nil {
		fmt.Printf("Error getting client configuration: %v\n", err)
		os.Exit(1)
	}
	tmpClientConfig := &config.ClientConfig{}
	if err = json.Unmarshal(clientConfigJson, tmpClientConfig); err != nil { // Unmarshal JSON into struct
		fmt.Printf("Error unmarshaling client configuration JSON: %v\n", err)
		os.Exit(1)
	}
	err = config.InitializeClientConfig(tmpClientConfig)
	if err != nil {
		fmt.Printf("Error initializing client configuration: %v\n", err)
		os.Exit(1)
	}

	// Verify client configuration loaded correctly
	clientConfig := config.GetClientConfig()
	if clientConfig == nil {
		fmt.Printf("Client configuration is nil after initialization\n")
		os.Exit(1)
	}
	fmt.Printf("Client configuration loaded successfully\n")

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

	// Update hardware information in client data
	cd := &config.ClientData{}
	if err := config.InitializeClientData(cd); err != nil {
		fmt.Printf("Error initializing client data: %v\n", err)
		os.Exit(1)
	}
	getClientData()

	clientData := config.GetClientData()
	if clientData == nil {
		fmt.Printf("Client data is nil after initialization\n")
		os.Exit(1)
	}
	fmt.Printf("Network interfaces detected: %d\n", len(clientData.Hardware.Network))
	for ifName, netData := range clientData.Hardware.Network {
		macAddr := netData.MACAddress
		ipAddr := netData.IPAddress
		linkUp := "down"
		if netData.NetworkLinkUp != nil && *netData.NetworkLinkUp {
			linkUp = "up"
		}
		fmt.Printf("Interface: %s, MAC: %s, IP: %s, Link: %s\n", ifName, macAddr, ipAddr, linkUp)
	}
	fmt.Printf("\nUIT Client setup completed successfully.\n")

	os.Exit(0)
}
