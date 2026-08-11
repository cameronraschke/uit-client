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

func GetCPUData() (usagePcnt float64, mhz float64, temp float64, err error) {

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
				fmt.Printf("col count=%d", len(cols))
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
				fmt.Printf("user=%d\nnice=%d\nsystem=%d\nidle=%d\niowait=%d\nirq=%d\nsoftirq=%d\nsteal=%d\nguest=%d\nguest_nice=%d\n", user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice)
				totalCPUTime = user + nice + system + idle + iowait + irq + softirq + steal + guest + guest_nice
				fmt.Printf("Total CPU Time=%d jiffies\n", totalCPUTime)
				idleTime := idle + iowait
				fmt.Printf("Total Idle Time=%d jiffies\n", idleTime)
				totalActiveCPUTime = totalCPUTime - idleTime
				fmt.Printf("Active Time=%d jiffies\n", totalActiveCPUTime)
			}
		}
		return totalActiveCPUTime, totalCPUTime
	}

	active1, total1 := processProcStat()
	fmt.Printf("active1=%d, total1=%d\n", active1, total1)
	time.Sleep(1 * time.Second)
	active2, total2 := processProcStat()
	fmt.Printf("active2=%d, total2=%d\n", active2, total2)
	usagePcnt = ((float64(active2) - float64(active1)) / (float64(total2) - float64(total1))) * 100
	fmt.Printf("Usage after sleep=%.4f\n", usagePcnt)

	return usagePcnt, mhz, temp, nil
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
