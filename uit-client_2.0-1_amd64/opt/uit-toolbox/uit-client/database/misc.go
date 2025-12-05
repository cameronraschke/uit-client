package database

import (
	"fmt"
	"os"
	"time"
)

func GetSystemTime() time.Time {
	return time.Now()
}

func GetUptime() (time.Duration, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	var uptimeSeconds float64
	_, err = fmt.Sscanf(string(data), "%f", &uptimeSeconds)
	if err != nil {
		return 0, err
	}
	return time.Duration(uptimeSeconds * float64(time.Second)), nil
}
