package hardware

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/cpu"
)

func readFileAndTrim(filePath string) string {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(fileBytes))
}

func readUint(filePath string) uint64 {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(fileBytes)), 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func readUintBool(filePath string) bool {
	return readUint(filePath) == 1
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func IsARM64() bool {
	return false
}

func IsX86_64() bool {
	return cpu.X86.HasSSE2
}

func GetSystemSerial() (string, error) {
	data := readFileAndTrim("/sys/class/dmi/id/product_serial")
	if data == "" {
		return "", fmt.Errorf("system serial does not exist")
	}
	return string(data), nil
}

func GetSystemUUID() (string, error) {
	data := readFileAndTrim("/sys/class/dmi/id/product_uuid")
	if data == "" {
		return "", fmt.Errorf("system UUID does not exist")
	}
	return string(data), nil
}

func GetSystemVendor() (string, error) {
	data := readFileAndTrim("/sys/class/dmi/id/sys_vendor")
	if data == "" {
		return "", fmt.Errorf("system vendor does not exist")
	}
	return string(data), nil
}

func GetSystemSKU() string {
	return string(readFileAndTrim("/sys/class/dmi/id/product_sku"))
}
