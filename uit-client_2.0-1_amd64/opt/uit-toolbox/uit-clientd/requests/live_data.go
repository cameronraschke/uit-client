package requests

import (
	"fmt"
	"io"
	"iter"
	"os"
	"strconv"
	"strings"
	"time"
)

type LiveDataRequest struct {
	Hardware   *HardwareDataRequest `json:"hardware"`
	Job        *JobQueueDataRequest `json:"job"`
	Status     *AppStatusRequest    `json:"status"`
	Screenshot []byte               `json:"live_screenshot"`
}

type HardwareDataRequest struct {
	BatteryChargePcnt float64 `json:"battery_charge_pcnt"`
	BatteryStatus     string  `json:"battery_status"`
	CPUUsagePcnt      float64 `json:"cpu_usage_pcnt"`
	CPUMhz            int64   `json:"cpu_mhz"`
	CPUTemp           float64 `json:"cpu_temp"`
	MemUsageKB        int64   `json:"int64"`
	MemCapacityKB     int64   `json:"mem_capacity_kb"`
	NetLinkSpeedMbit  float64 `json:"net_link_speed_mbit"`
	NetUsageMbit      float64 `json:"net_usage_mbit"`
	PowerUsageWatts   float64 `json:"power_usage_watts"`
}

type JobQueueDataRequest struct {
	CloneMode     string `json:"clone_mode"`
	EraseMode     string `json:"erase_mode"`
	IsQueued      *bool  `json:"is_queued"`
	IsRunning     *bool  `json:"is_running"`
	Name          string `json:"name"`
	NameFormatted string `json:"name_formatted"`
	DiskImageName string `json:"disk_image_name"`
	QueuePosition string `json:"queue_position"`
}

type AppStatusRequest struct {
	AppUptime     int64         `json:"app_uptime"`
	Current       string        `json:"current"`
	IsOnline      *bool         `json:"is_online"`
	IsPluggedIn   *bool         `json:"is_plugged_in"`
	KernelUpdated *bool         `json:"kernel_updated"`
	LastHeard     time.Time     `json:"last_heard"`
	SystemUptime  time.Duration `json:"system_uptime"`
}

func GetCPUData() (cpuUsagePcnt float64, cpuMHzAvg float64, cpuTemp float64, err error) {

	readProcStat := func() (iter.Seq[string], error) {
		f, err := os.Open("/proc/stat")
		if err != nil {
			return nil, fmt.Errorf("Error opening file '/proc/stat': %v", err)
		}
		defer f.Close()

		data := make([]byte, 1024) // We only need the first line
		count, err := f.Read(data)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading /proc/stat")
		}

		return strings.Lines(string(data[:count])), nil
	}

	processProcStat := func() (totalActiveCPUTime int64, totalCPUTime int64) {
		lines, _ := readProcStat()
		for i := range lines {
			// The space after cpu is important, matches only first aggregated row
			if strings.HasPrefix(i, "cpu ") {
				cols := strings.Fields(i)
				user, _ := strconv.ParseInt(cols[1], 10, 64)
				nice, _ := strconv.ParseInt(cols[2], 10, 64)
				system, _ := strconv.ParseInt(cols[3], 10, 64)
				idle, _ := strconv.ParseInt(cols[4], 10, 64)
				iowait, _ := strconv.ParseInt(cols[5], 10, 64)
				irq, _ := strconv.ParseInt(cols[6], 10, 64)
				softirq, _ := strconv.ParseInt(cols[7], 10, 64)
				steal, _ := strconv.ParseInt(cols[8], 10, 64)
				guest, _ := strconv.ParseInt(cols[9], 10, 64)
				guest_nice, _ := strconv.ParseInt(cols[10], 10, 64)
				totalCPUTime = user + nice + system + idle + iowait + irq + softirq + steal + guest + guest_nice
				idleTime := idle + iowait
				totalActiveCPUTime = totalCPUTime - idleTime
			}
		}
		return totalActiveCPUTime, totalCPUTime
	}

	// CPU usage percent
	active1, total1 := processProcStat()
	time.Sleep(1 * time.Second)
	active2, total2 := processProcStat()
	cpuUsagePcnt = ((float64(active2) - float64(active1)) / (float64(total2) - float64(total1))) * 100

	// CPU MHz
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error opening /proc/cpuinfo: %v", err)
	}

	lines := strings.Lines(string(data))
	var totalMHz float64
	var entryCount int
	for line := range lines {
		if strings.HasPrefix(line, "cpu MHz") {
			fieldsArr := strings.Fields(line)
			mhz, err := strconv.ParseFloat(fieldsArr[3], 64)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("error parsing /proc/cpuinfo columns/fields: %v", err)
			}
			totalMHz = totalMHz + mhz
			entryCount++
		}
	}

	cpuMHzAvg = totalMHz / float64(entryCount)

	return cpuUsagePcnt, cpuMHzAvg, cpuTemp, nil
}

func GetMemoryData() (totalCapacityKB int64, totalUsageKB int64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, fmt.Errorf("Error opening file '/proc/meminfo': %v", err)
	}
	defer f.Close()

	data := make([]byte, 2048) // It's around 1500 KB
	count, err := f.Read(data)
	if err != nil {
		return 0, 0, fmt.Errorf("error reading /proc/meminfo")
	}

	lines := strings.Lines(string(data[:count]))

	for i := range lines {
		if totalCapacityKB > 0 && totalUsageKB > 0 {
			break
		}
		if strings.HasPrefix(i, "MemTotal") {
			memTotalLine := strings.Fields(i)
			if len(memTotalLine) == 0 {
				return 0, 0, fmt.Errorf("Error parsing MemTotal in '/proc/meminfo': %s", "memTotalLine has len of 0")
			}
			totalCapacityKB, err = strconv.ParseInt(memTotalLine[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("Error parsing MemTotal in '/proc/meminfo': %v", err)
			}
		}
		if strings.HasPrefix(i, "MemAvailable") {
			memAvailableLine := strings.Fields(i)
			if len(memAvailableLine) == 0 {
				return 0, 0, fmt.Errorf("Error parsing MemAvailable in '/proc/meminfo': %s", "memAvailableLine has len of 0")
			}
			totalUsageKB, err = strconv.ParseInt(memAvailableLine[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("Error parsing MemAvailable in '/proc/meminfo': %v", err)
			}
		}
	}
	return totalCapacityKB, totalUsageKB, nil
}
