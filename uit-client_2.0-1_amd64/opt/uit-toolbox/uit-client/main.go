//go:build linux && amd64

package main

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"uitclient/api"
	"uitclient/cli"
	"uitclient/client"
	"uitclient/config"
	"uitclient/types"

	"golang.org/x/sys/cpu"
	"golang.org/x/sys/unix"
)

const clearScreen = `\e[1;1H\e[2J`

// const clearScreen = "\033[H\033[2J"

func getClientData() {
	// System data
	serialPtr, err := client.GetSystemSerial()
	if err != nil {
		fmt.Printf("Error getting system serial: %v\n", err)
		os.Exit(1)
	}
	if serialPtr == nil {
		fmt.Printf("System serial number is empty\n")
		os.Exit(1)
	}
	config.SetSystemSerial(serialPtr)

	clientLookup, err := api.SerialLookup(*serialPtr)
	if err != nil {
		fmt.Printf("Error looking up serial number: %v\n", err)
		os.Exit(1)
	}
	if clientLookup.Tagnumber == nil || *clientLookup.Tagnumber < 1 {
		fmt.Fprintln(os.Stderr, "Invalid tagnumber retrieved")
		os.Exit(1)
	}
	config.SetTagnumber(clientLookup.Tagnumber)

	systemUUID, err := client.GetSystemUUID()
	if err != nil {
		fmt.Printf("Error getting system UUID: %v\n", err)
		os.Exit(1)
	}
	if systemUUID == nil || *systemUUID == "" {
		fmt.Printf("System UUID is empty\n")
		os.Exit(1)
	}
	config.SetSystemUUID(systemUUID)

	manufacturer := client.GetSystemManufacturer()
	config.SetManufacturer(manufacturer)

	model := client.GetSystemModel()
	config.SetModel(model)

	sku := client.GetSystemSKU()
	config.SetSKU(sku)

	// Network data
	connectedToHost := true // assumed, got client config from server
	config.SetConnectedToHost(&connectedToHost)

	networkInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Error getting network interfaces: %v\n", err)
		os.Exit(1)
	}

	networkMap := make(map[string]types.NetworkHardwareData)
	for _, netIf := range networkInterfaces {
		ifName := netIf.Name
		if ifName == "" {
			fmt.Printf("Network interface has no name, skipping\n")
			continue
		}
		macAddress := netIf.HardwareAddr.String()
		macAddressPtr := &macAddress
		if *macAddressPtr == "" { // MAC address cannot be nil here, check for empty string instead
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
		networkMap[ifName] = types.NetworkHardwareData{
			MACAddress:    macAddressPtr,
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

	config.UpdateClientData(func(clientData *types.ClientData) {
		clientData.Serial = serialPtr
	})
}

func main() {
	recover()
	// Initial startup, checks, and configuration loading

	// Clear the terminal screen
	fmt.Printf(clearScreen)
	fmt.Printf("Starting UIT Client...\n\n")

	term, err := cli.InitTerminalWithRaw(true)
	if err != nil {
		fmt.Printf("Error initializing terminal: %v\n", err)
		os.Exit(1)
	}

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
	clientConfigJson, err := api.GetClientConfig()
	if err != nil {
		fmt.Printf("Error getting client configuration: %v\n", err)
		os.Exit(1)
	}
	err = config.InitializeClientConfig(clientConfigJson)
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

	serialPtr, err := client.GetSystemSerial()
	if err != nil {
		fmt.Printf("Error getting system serial: %v\n", err)
		os.Exit(1)
	}
	if serialPtr == nil || *serialPtr == "" {
		fmt.Printf("System serial number is empty\n")
		os.Exit(1)
	}

	clientLookup, err := api.SerialLookup(*serialPtr)
	if err != nil {
		fmt.Printf("Error looking up serial number: %v\n", err)
		os.Exit(1)
	}
	if clientLookup.Tagnumber == nil || *clientLookup.Tagnumber < 1 {
		fmt.Printf("Invalid tagnumber retrieved: %d\n", *clientLookup.Tagnumber)
		os.Exit(1)
	}

	fmt.Printf("Tagnumber: %d, System Serial: %s\n", *clientLookup.Tagnumber, *serialPtr)

	hasFP := cpu.X86.HasAES || cpu.ARM64.HasSHA1 || cpu.ARM64.HasSHA2 || cpu.ARM64.HasSHA3 || cpu.ARM64.HasCRC32
	if hasFP {
		fmt.Printf("CPU has encryption acceleration\n")
	}

	devicePath, totalDevices, err := term.SelectBlockDevices()
	if err != nil {
		fmt.Printf("Error selecting block device: %v\n", err)
		os.Exit(1)
	}
	if totalDevices <= 0 {
		fmt.Printf("No block devices found on system\n")
		os.Exit(1)
	}

	fmt.Printf("Selected block device path: %s\n", devicePath)

	// Update client information in client data
	cd := &types.ClientData{}
	if err := config.InitializeClientData(cd); err != nil {
		fmt.Printf("Error initializing client data: %v\n", err)
		os.Exit(1)
	}
	getClientData()

	clientData := config.GetClientData()
	if clientData == (types.ClientData{}) {
		fmt.Printf("Client data is nil after initialization\n")
		os.Exit(1)
	}
	fmt.Printf("Network interfaces detected: %d\n", len(clientData.Hardware.Network))
	for ifName, netData := range clientData.Hardware.Network {
		macAddrPtr := netData.MACAddress
		ipAddr := netData.IPAddress
		linkUp := "down"
		if netData.NetworkLinkUp != nil && *netData.NetworkLinkUp {
			linkUp = "up"
		}
		fmt.Printf("Interface: %s, MAC: %s, IP: %s, Link: %s\n", ifName, *macAddrPtr, ipAddr, linkUp)
	}
	fmt.Printf("\nUIT Client setup completed successfully.\n")

	os.Exit(0)
}
