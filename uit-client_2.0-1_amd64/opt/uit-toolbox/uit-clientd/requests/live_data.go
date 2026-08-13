package requests

import (
	"context"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	hwmonDirRoot     = "/sys/class/hwmon/"
	powercapRootDir  = "/sys/class/powercap/"
	powercapRootDir2 = "/sys/class/powercap/intel-rapl/"
	netIfRootDir     = "/sys/class/net/"
)

type LiveDataRequest struct {
	Hardware   *HardwareDataRequest `json:"hardware"`
	Job        *JobQueueDataRequest `json:"job"`
	Status     *AppStatusRequest    `json:"status"`
	Screenshot []byte               `json:"live_screenshot"`
}

type HardwareDataRequest struct {
	BatteryChargePcnt int64   `json:"battery_charge_pcnt"`
	BatteryStatus     string  `json:"battery_status"`
	CPUUsagePcnt      float64 `json:"cpu_usage_pcnt"`
	CPUMhz            float64 `json:"cpu_mhz"`
	CPUTemp           float64 `json:"cpu_temp"`
	DiskTemp          float64 `json:"disk_temp"`
	DiskMaxTemp       float64 `json:"disk_max_temp"`
	MemUsageKB        int64   `json:"memory_usage_kb"`
	MemCapacityKB     int64   `json:"memory_capacity_kb"`
	NetLinkSpeedMbit  int64   `json:"net_link_speed_mbit"`
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

func GetCPUData(rootCtx context.Context) (cpuUsagePcnt float64, cpuMHzAvg float64, cpuTemp float64, err error) {
	ctx, ctxCancel := context.WithCancel(rootCtx)
	defer ctxCancel()
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	var errChanOnce sync.Once // Only get first error
	sendToErrChanOnce := func(err error) {
		if err == nil {
			return
		}
		errChanOnce.Do(func() {
			errChan <- err
			ctxCancel()
		})
	}

	readProcStat := func(ctx context.Context) (iter.Seq[string], error) {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		f, err := os.Open("/proc/stat")
		if err != nil {
			return nil, fmt.Errorf("Error opening file '/proc/stat': %w", err)
		}
		defer f.Close()

		data := make([]byte, 1024) // We only need the first line
		count, err := f.Read(data)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading /proc/stat")
		}

		return strings.Lines(string(data[:count])), nil
	}

	processProcStat := func(ctx context.Context) (totalActiveCPUTime int64, totalCPUTime int64, err error) {
		lines, err := readProcStat(ctx)
		if err != nil {
			return 0, 0, fmt.Errorf("context error: %w", ctx.Err())
		}
		for i := range lines {
			if ctx.Err() != nil {
				return 0, 0, fmt.Errorf("context error: %w", ctx.Err())
			}
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
		return totalActiveCPUTime, totalCPUTime, nil
	}

	// CPU usage percent
	wg.Go(func() {
		active1, total1, err := processProcStat(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error processing first read of /proc/stat: %w", err))
			return
		}
		timer := time.NewTimer(1 * time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
			return
		case <-timer.C:
			// continue
		}
		active2, total2, err := processProcStat(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error processing second read of /proc/stat: %w", err))
			return
		}
		activeDelta := float64(active2) - float64(active1)
		totalDelta := float64(total2) - float64(total1)
		if activeDelta == totalDelta || activeDelta == 0 || totalDelta == 0 {
			sendToErrChanOnce(fmt.Errorf("arithmetic error between active and total CPU time deltas"))
			return
		}

		cpuUsagePcnt = (activeDelta / totalDelta) * 100
	})

	// CPU MHz
	wg.Go(func() {
		data, err := os.ReadFile("/proc/cpuinfo")
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error opening /proc/cpuinfo: %w", err))
			return
		}

		lines := strings.Lines(string(data))
		var totalMHz float64
		var entryCount int
		for line := range lines {
			if ctx.Err() != nil {
				sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
				return
			}
			if strings.HasPrefix(line, "cpu MHz") {
				fieldsArr := strings.Fields(line)
				mhz, err := strconv.ParseFloat(fieldsArr[3], 64)
				if err != nil {
					sendToErrChanOnce(fmt.Errorf("error parsing /proc/cpuinfo columns/fields: %w", err))
					return
				}
				totalMHz = totalMHz + mhz
				entryCount++
			}
		}

		cpuMHzAvg = totalMHz / float64(entryCount)
	})

	// CPU temp
	wg.Go(func() {
		// find coretemp name/type
		hwmons, err := os.ReadDir("/sys/class/hwmon/")
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error opening directory /sys/class/hwmon: %w", err))
			return
		}

		var totalDegrees int64
		var totalEntries int64
		for _, hwmonDir := range hwmons {
			if ctx.Err() != nil {
				sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
				return
			}
			hwmonNamePath := filepath.Join("/sys/class/hwmon/", hwmonDir.Name(), "name")
			data, err := os.ReadFile(hwmonNamePath)
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("cannot open hwmon file '%s': %w", hwmonNamePath, err))
				return
			}
			hwmonName := strings.TrimSpace(string(data)) // string not trimmed for some reason
			if hwmonName != "coretemp" {
				continue
			}
			hwmonDir := filepath.Join("/sys/class/hwmon/", hwmonDir.Name())
			hwmonDirEntries, err := os.ReadDir(hwmonDir)
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("cannot open hwmon dir: %w", err))
				return
			}
			if len(hwmonDirEntries) == 0 {
				sendToErrChanOnce(fmt.Errorf("hwmonDirEntries has len of 0"))
				return
			}
			for _, input := range hwmonDirEntries {
				if ctx.Err() != nil {
					sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
					return
				}
				matches, err := regexp.MatchString(`temp[0-9]+\_input`, input.Name())
				if err != nil || !matches {
					continue
				}
				fullFilePath := filepath.Join(hwmonDir, input.Name())
				fileBytes, err := os.ReadFile(fullFilePath)
				if err != nil {
					sendToErrChanOnce(fmt.Errorf("error reading temp input value from '%s': %w", input.Name(), err))
					return
				}
				intVal, err := strconv.ParseInt(strings.TrimSpace(string(fileBytes)), 10, 64)
				if err != nil {
					sendToErrChanOnce(fmt.Errorf("cannot parse hwmon temp value for file '%s': %w", fullFilePath, err))
					return
				}
				totalDegrees = totalDegrees + intVal
				totalEntries++
			}
		}
		cpuTemp = (float64(totalDegrees) / float64(totalEntries)) / 1000
	})

	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		return 0, 0, 0, fmt.Errorf("error getting CPU data: %w", err)
	}

	return cpuUsagePcnt, cpuMHzAvg, cpuTemp, nil
}

func GetMemoryData() (totalCapacityKB int64, totalUsageKB int64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, fmt.Errorf("Error opening file '/proc/meminfo': %w", err)
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
				return 0, 0, fmt.Errorf("Error parsing MemTotal in '/proc/meminfo': %w", err)
			}
		}
		if strings.HasPrefix(i, "MemAvailable") {
			memAvailableLine := strings.Fields(i)
			if len(memAvailableLine) == 0 {
				return 0, 0, fmt.Errorf("Error parsing MemAvailable in '/proc/meminfo': %s", "memAvailableLine has len of 0")
			}
			memAvailable, err := strconv.ParseInt(memAvailableLine[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("Error parsing MemAvailable in '/proc/meminfo': %w", err)
			}
			totalUsageKB = totalCapacityKB - memAvailable
		}
	}
	return totalCapacityKB, totalUsageKB, nil
}

func GetPowerSupplyData(rootCtx context.Context) (powerUsageWatts float64, batChargePcnt int64, batStatus string, err error) {
	ctx, ctxCancel := context.WithCancel(rootCtx)
	defer ctxCancel()
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	var errChanOnce sync.Once // Only get first error
	sendToErrChanOnce := func(err error) {
		if err == nil {
			return
		}
		errChanOnce.Do(func() {
			errChan <- err
			ctxCancel()
		})
	}

	// Returns total watts used by system (at least that is reported by the kernel)
	getTotalWatts := func(ctx context.Context) (float64, error) {
		powercapDirs, err := os.ReadDir(powercapRootDir)
		if err != nil {
			return 0, fmt.Errorf("cannot open powercap directory: %w", err)
		}
		var totaluJoules int64
		for _, dir := range powercapDirs {
			if ctx.Err() != nil {
				return 0, fmt.Errorf("context error: %w", ctx.Err())
			}
			powercapDir := filepath.Join(powercapRootDir, dir.Name(), "energy_uj")
			powercapDir2 := filepath.Join(powercapRootDir2, "intel-rapl:0", "energy_uj")
			var powercapBytes []byte
			var err1 error
			var err2 error
			powercapBytes, err1 = os.ReadFile(powercapDir)
			if err1 != nil { // fallback to other directory (powercapDir2)
				powercapBytes, err2 = os.ReadFile(powercapDir2)
				if err2 != nil {
					return 0, fmt.Errorf("cannot read '%s' (%w) or '%s' (%w)", powercapDir, err1, powercapDir2, err2)
				}
			}
			powercapStr := strings.TrimSpace(string(powercapBytes))
			joules, err := strconv.ParseInt(powercapStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("unable to parse total microjoules value: %w", err)
			}
			totaluJoules += joules
		}
		if totaluJoules == 0 {
			return 0, fmt.Errorf("aggregate microjoule value is zero")
		}
		return (float64(totaluJoules) / 1e6), nil
	}

	// Current wattage
	wg.Go(func() {
		totalWatts1, err := getTotalWatts(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error during wattage reading: %w", err))
			return
		}
		timer := time.NewTimer(1 * time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
			return
		case <-timer.C:
			// continue
		}
		totalWatts2, err := getTotalWatts(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error during wattage reading: %w", err))
			return
		}
		powerUsageWatts = totalWatts2 - totalWatts1
	})

	// Battery charge percent and status string
	wg.Go(func() {
		hwmonDirs, err := os.ReadDir(hwmonDirRoot)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error opening directory '%s': %w", hwmonDirRoot, err))
			return
		}
		for _, dir := range hwmonDirs {
			if ctx.Err() != nil {
				sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
				return
			}
			hwmonDir := filepath.Join(hwmonDirRoot, dir.Name())
			deviceNameBytes, _ := os.ReadFile(filepath.Join(hwmonDir, "name"))
			deviceName := strings.TrimSpace(string(deviceNameBytes))
			if deviceName != "BAT0" && deviceName != "BAT1" {
				continue
			}
			chargePcntBytes, err := os.ReadFile(filepath.Join(hwmonDir, "device", "capacity"))
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("cannot read file '%s': %w", filepath.Join(hwmonDir, "device", "capacity"), err))
				return
			}
			chargePcntStr := strings.TrimSpace(string(chargePcntBytes))
			batChargePcnt, err = strconv.ParseInt(chargePcntStr, 10, 64)
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("cannot parse battery charge percent: %w", err))
				return
			}

			statusBytes, _ := os.ReadFile(filepath.Join(hwmonDir, "device", "status"))
			batStatus = strings.TrimSpace(string(statusBytes))
		}
	})

	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		return 0, 0, "", fmt.Errorf("error getting power supply data: %w", err)
	}
	return powerUsageWatts, batChargePcnt, batStatus, nil
}

func GetDiskData() (curTemp float64, maxTemp float64, err error) {
	hwmonDirs, err := os.ReadDir(hwmonDirRoot)
	if err != nil {
		return 0, 0, fmt.Errorf("error opening directory '%s': %w", hwmonDirRoot, err)
	}
	for _, dir := range hwmonDirs {
		hwmonDir := filepath.Join(hwmonDirRoot, dir.Name())
		hwmonNameBytes, _ := os.ReadFile(filepath.Join(hwmonDir, "name"))
		hwmonNameStr := strings.TrimSpace(string(hwmonNameBytes))
		if hwmonNameStr != "nvme" {
			continue
		}

		// TODO: loop over temp*_input and temp1*_max
		curTempBytes, _ := os.ReadFile(filepath.Join(hwmonDir, "temp1_input"))
		curTempStr := strings.TrimSpace(string(curTempBytes))
		cur, err := strconv.ParseInt(curTempStr, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot parse current disk temp: %w", err)
		}
		curTemp = float64(cur) / 1000

		maxTempBytes, err := os.ReadFile(filepath.Join(hwmonDir, "temp1_max"))
		if err != nil {
			return 0, 0, fmt.Errorf("cannot read file '%s': %w", filepath.Join(hwmonDir, "temp1_max"), err)
		}
		maxTempStr := strings.TrimSpace(string(maxTempBytes))
		max, err := strconv.ParseInt(maxTempStr, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot parse max disk temp: %w", err)
		}
		maxTemp = float64(max) / 1000
	}
	return curTemp, maxTemp, nil
}

func GetNetworkData(rootCtx context.Context) (linkSpeed int64, mbpsThroughput float64, err error) {
	var initialValidNetDirs []string
	var initialValidNetDirsMu sync.RWMutex
	ctx, ctxCancel := context.WithCancel(rootCtx)
	defer ctxCancel()
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	var errChanOnce sync.Once // Only get first error
	sendToErrChanOnce := func(err error) {
		if err == nil {
			return
		}
		errChanOnce.Do(func() {
			errChan <- err
			ctxCancel()
		})
	}

	initValidNetIFDirs := func(ctx context.Context) error {
		netDirs, err := os.ReadDir(netIfRootDir)
		if err != nil {
			return fmt.Errorf("error reading dir '%s': %w", netIfRootDir, err)
		}
		for _, dir := range netDirs {
			if ctx.Err() != nil {
				return fmt.Errorf("context error: %w", ctx.Err())
			}
			ifDir := filepath.Join(netIfRootDir, dir.Name())
			b, err := os.ReadFile(filepath.Join(ifDir, "type"))
			if err != nil {
				return fmt.Errorf("cannot read interface type of '%s': %w", dir.Name(), err)
			}
			bStr := strings.TrimSpace(string(b))
			ifType, err := strconv.ParseInt(bStr, 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse interface type of '%s' to int64: %w", dir.Name(), err)
			}
			// interface types in linux kernel: include/uapi/linux/if_arp.h
			if ifType != 1 || ifType == 772 { // redundant, but 772 is a loopback device
				continue
			}
			initialValidNetDirsMu.Lock()
			initialValidNetDirs = append(initialValidNetDirs, ifDir)
			initialValidNetDirsMu.Unlock()
		}
		return nil
	}
	// Try to have this write occur before goroutines or anything else writes to it
	var populateValidDirsOnce sync.Once
	populateValidDirs := func() {
		populateValidDirsOnce.Do(func() {
			if err := initValidNetIFDirs(ctx); err != nil {
				sendToErrChanOnce(fmt.Errorf("error during initValidNetIFDirs: %w", err))
				return
			}
		})
	}
	populateValidDirs()

	// returns slice of directories ex /sys/class/net/en...
	getValidNetIFDirs := func(ctx context.Context) ([]string, error) {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context error: %w", ctx.Err())
		}
		initialValidNetDirsMu.RLock()
		defer initialValidNetDirsMu.RUnlock()
		if len(initialValidNetDirs) == 0 {
			return nil, fmt.Errorf("initialValidNetDirs has len of 0")
		}
		validNetDirsCopy := make([]string, len(initialValidNetDirs))
		copy(validNetDirsCopy, initialValidNetDirs)
		return validNetDirsCopy, nil
	}

	// gets sum of /sys/class/net/*/statistics/(tx|rx)_bytes - only for valid ethernet interfaces
	getTxRxSum := func(ctx context.Context) (int64, error) {
		var txRxSum int64
		allValidNetDirs, err := getValidNetIFDirs(ctx)
		if err != nil {
			return 0, fmt.Errorf("error retrieving all valid network interface directories: %w", err)
		}
		// before allValidNetDirs is returned, it checks to make sure it's populated
		for _, dir := range allValidNetDirs {
			if ctx.Err() != nil {
				return 0, fmt.Errorf("context error: %w", ctx.Err())
			}
			txTotalBytes, err := os.ReadFile(filepath.Join(dir, "statistics", "tx_bytes"))
			if err != nil {
				return 0, fmt.Errorf("cannot read tx_bytes of interface '%s': %w", dir, err)
			}
			txTotalStr := strings.TrimSpace(string(txTotalBytes))
			txTotal, err := strconv.ParseInt(txTotalStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse tx_bytes of '%s' to int64", filepath.Base(dir))
			}
			rxTotalBytes, err := os.ReadFile(filepath.Join(dir, "statistics", "rx_bytes"))
			if err != nil {
				return 0, fmt.Errorf("cannot read rx_bytes of interface '%s': %w", filepath.Base(dir), err)
			}
			rxTotalStr := strings.TrimSpace(string(rxTotalBytes))
			rxTotal, err := strconv.ParseInt(rxTotalStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot parse rx_bytes of '%s' to int64", filepath.Base(dir))
			}
			txRxSum += txTotal + rxTotal
		}
		return txRxSum, nil
	}

	// TX and RX sum calculations
	wg.Go(func() {
		txrxSum1, err := getTxRxSum(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error reading first tx/rx sum value: %w", err))
			return
		}
		timer := time.NewTimer(1 * time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
			return
		case <-timer.C:
			// continue
		}
		txrxSum2, err := getTxRxSum(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error reading second tx/rx sum value: %w", err))
			return
		}
		mbpsThroughput = (float64(txrxSum2) - float64(txrxSum1)) * 8 / 1e6
	})

	// Link speed
	wg.Go(func() {
		// before allValidNetDirs is returned, it checks to make sure it's populated
		allValidNetDirs, err := getValidNetIFDirs(ctx)
		if err != nil {
			sendToErrChanOnce(fmt.Errorf("error retrieving all valid network interface directories: %w", err))
			return
		}
		for _, dir := range allValidNetDirs {
			if ctx.Err() != nil {
				sendToErrChanOnce(fmt.Errorf("context error: %w", ctx.Err()))
				return
			}
			pluggedInBytes, err := os.ReadFile(filepath.Join(dir, "carrier"))
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("error reading link speed of interface '%s'", filepath.Base(dir)))
				return
			}
			pluggedInStr := strings.TrimSpace(string(pluggedInBytes))
			if pluggedInStr != "1" { // means it's active or plugged in, no need to convert to int really
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, "speed"))
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("error reading link speed of interface '%s'", filepath.Base(dir)))
				return
			}
			bStr := strings.TrimSpace(string(b))
			ls, err := strconv.ParseInt(bStr, 10, 64)
			if err != nil {
				sendToErrChanOnce(fmt.Errorf("error parsing net interface link speed of '%s': %w", filepath.Base(dir), err))
				return
			}
			linkSpeed = ls
		}
	})

	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		return 0, 0, err
	}
	return linkSpeed, mbpsThroughput, nil
}

func printHardwareData() {
	parentCtx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	var wg sync.WaitGroup
	errChan := make(chan error, 1)
	sendToErrChan := func(err error) {
		if err != nil {
			errChan <- err
			ctxCancel()
		}
	}

	hardwareData := new(HardwareDataRequest)

	// CPU data
	wg.Go(func() {
		cpuUsage, cpuMHz, cpuTemp, err := GetCPUData(parentCtx)
		if err != nil {
			sendToErrChan(fmt.Errorf("error retrieving CPU data: %w", err))
			return
		}
		hardwareData.CPUUsagePcnt = cpuUsage
		hardwareData.CPUMhz = cpuMHz
		hardwareData.CPUTemp = cpuTemp
	})

	// Memory data
	wg.Go(func() {
		memCapacity, memUsage, err := GetMemoryData()
		if err != nil {
			sendToErrChan(fmt.Errorf("error retrieving memory data: %w", err))
			return
		}
		hardwareData.MemCapacityKB = memCapacity
		hardwareData.MemUsageKB = memUsage
	})

	// Power supply and battery data
	wg.Go(func() {
		powerUsage, batCharge, batStatus, err := GetPowerSupplyData(parentCtx)
		if err != nil {
			sendToErrChan(fmt.Errorf("error retrieving battery data: %w", err))
			return
		}
		hardwareData.PowerUsageWatts = powerUsage
		hardwareData.BatteryChargePcnt = batCharge
		hardwareData.BatteryStatus = batStatus
	})

	// Disk data
	wg.Go(func() {
		diskTemp, diskMaxTemp, err := GetDiskData()
		if err != nil {
			sendToErrChan(fmt.Errorf("error retrieving disk data: %w", err))
			return
		}
		hardwareData.DiskTemp = diskTemp
		hardwareData.DiskMaxTemp = diskMaxTemp
	})

	wg.Go(func() {
		netLinkSpeed, netThroughput, err := GetNetworkData(parentCtx)
		if err != nil {
			sendToErrChan(fmt.Errorf("error retrieving net interface data: %w\n", err))
			return
		}
		hardwareData.NetLinkSpeedMbit = netLinkSpeed
		hardwareData.NetUsageMbit = netThroughput
	})

	wg.Wait()
	close(errChan)

	for err := range errChan {
		fmt.Fprintf(os.Stderr, "[Error] hardware data error: %v", err)
	}

	fmt.Printf("data: \n%#v\n", hardwareData)
}
