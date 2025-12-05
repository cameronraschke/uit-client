//go:build linux && amd64

package client

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

func readFileAndTrim(filePath string) *string {
	if strings.TrimSpace(filePath) == "" {
		return nil
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	if fileBytes == nil {
		return nil
	}
	if len(fileBytes) == 0 {
		return nil
	}
	trimmed := strings.TrimSpace(string(fileBytes))
	return &trimmed
}

func readUintPtr(filePath string) *int64 {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	if fileBytes == nil {
		return nil
	}
	if len(fileBytes) == 0 {
		return nil
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(fileBytes)), 10, 64)
	if err != nil {
		return nil
	}
	if value == 0 {
		return nil
	}
	return &value
}

func readUintBool(filePath string) bool {
	// Sysfs boolean-like files typically contain "0" or "1".
	// Treat exactly "1" as true; anything else (including errors) as false.
	return readUintPtr(filePath) != nil && *readUintPtr(filePath) == 1
}

func readUintBoolPtr(filePath string) *bool {
	b, err := os.ReadFile(filePath)
	if err != nil || b == nil {
		return nil
	}
	s := strings.TrimSpace(string(b))
	switch s {
	case "1":
		v := true
		return &v
	case "0":
		v := false
		return &v
	}

	// Fallback: try parsing as uint
	if u, err := strconv.ParseUint(s, 10, 64); err == nil {
		if u == 1 {
			v := true
			return &v
		}
		if u == 0 {
			v := false
			return &v
		}
	}
	return nil
}

func fileExists(filePath string) bool {
	if filePath == "" || !utf8.ValidString(filePath) {
		return false
	}
	_, err := os.Stat(filePath)
	return err == nil
}

func GetSystemSerial() (*string, error) {
	data := readFileAndTrim("/sys/class/dmi/id/product_serial")
	if data == nil || *data == "" {
		return nil, fmt.Errorf("system serial does not exist")
	}
	return data, nil
}

func GetSystemUUID() (*string, error) {
	data := readFileAndTrim("/sys/class/dmi/id/product_uuid")
	if data == nil || *data == "" {
		return nil, fmt.Errorf("system UUID does not exist")
	}
	return data, nil
}

func GetSystemVendor() (*string, error) {
	data := readFileAndTrim("/sys/class/dmi/id/sys_vendor")
	if data == nil || *data == "" {
		return nil, fmt.Errorf("system vendor does not exist")
	}
	return data, nil
}

func GetSystemSKU() *string {
	data := readFileAndTrim("/sys/class/dmi/id/product_sku")
	if data == nil || *data == "" {
		return nil
	}
	return data
}
