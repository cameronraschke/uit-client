package main

import (
	"fmt"

	"github.com/microsoft/wmi"
)

func main() {
	fmt.Println("Collecting data for UIT Web")
	wmiInstance := wmi.WmiInstance{
		Class: "Win32_OperatingSystem",
	}
	win32Instance, err := wmi.NewWin32_OperatingSystemEx1(wmiInstance)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Win32_OperatingSystem instance:", win32Instance)
}
