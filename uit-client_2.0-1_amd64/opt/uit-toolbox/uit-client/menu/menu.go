//go:build linux && amd64

package menu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"uitclient/client"

	"golang.org/x/term"
)

const (
	termWidth  = 80
	termHeight = 33
)

var (
	termFd     = int(os.Stdin.Fd())
	menuReader = bufio.NewReader(os.Stdin)
	menuWriter = bufio.NewWriter(os.Stdout)
	rw         *bufio.ReadWriter
	t          *term.Terminal
)

func InitTerminal() (oldState *term.State, err error) {
	if term.IsTerminal(termFd) {
		oldState, err = term.MakeRaw(termFd)
		if err != nil {
			fmt.Printf("Error setting terminal to raw mode: %v\n", err)
			return nil, err
		}
	} else {
		fmt.Printf("Warning: Standard input is not a terminal\n")
	}
	curWidth, curHeight, err := term.GetSize(termFd)
	if err != nil {
		fmt.Printf("Error getting terminal size: %v\n", err)
		return nil, err
	}
	if curWidth < termWidth || curHeight < termHeight {
		fmt.Printf("Warning: Terminal size is smaller than recommended (%dx%d). Current size is %dx%d.\n",
			termWidth, termHeight, curWidth, curHeight)
	}
	rw = bufio.NewReadWriter(menuReader, menuWriter)
	menuWriter.Flush()
	t = term.NewTerminal(rw, "> ")
	if err := t.SetSize(termWidth, termHeight); err != nil {
		fmt.Printf("Error setting terminal size: %v\n", err)
		return nil, err
	}
	Echo("\nTerminal initialized with size %dx%d\n", termWidth, termHeight)
	return oldState, nil
}

func FlushMenu() {
	menuWriter.Flush()
}

func Echo(format string, a ...any) {
	fmt.Fprintf(t, format+"\n", a...)
	FlushMenu()
}

func Read(prompt string) (string, error) {
	t.SetPrompt(prompt)
	input, err := t.ReadLine()
	if err != nil {
		return "", err
	}
	t.SetPrompt("> ")
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("no input received")
	}
	return input, nil
}

func ReadOneChar(prompt string) (rune, error) {
	runeBuf := make([]rune, 1)
	input, err := Read(prompt)
	if err != nil {
		Echo("Error reading input: %v", err)
		return 0, err
	}
	runeBuf = []rune(input)
	if len(runeBuf) == 0 {
		return 0, fmt.Errorf("no input received")
	}
	return runeBuf[0], nil
}

func ReadOneCharLine(prompt string) (rune, error) {
	input, err := Read(prompt)
	if err != nil {
		Echo("Error reading input: %v", err)
		return 0, err
	}
	return []rune(input)[0], nil
}

func RestoreTerminal(oldState *term.State) {
	fmt.Printf("\nRestoring terminal settings...\n")
	if oldState != nil {
		term.Restore(termFd, oldState)
	}
}

func SelectBlockDevices() (string, int, error) {
	blockDevices, err := client.ListBlockDevices("/dev")
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
