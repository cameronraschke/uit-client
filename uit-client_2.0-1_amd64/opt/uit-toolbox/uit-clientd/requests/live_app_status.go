package requests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	appLastHeardFilePath = "/tmp/uit_app_last_heard"
	appUptimeFilePath    = "/tmp/uit_app_uptime"
)

var (
	appLastHeardBufMu sync.RWMutex
	appLastHeardBuf   bytes.Buffer
	appUptimeBufMu    sync.RWMutex
	appUptimeBuf      bytes.Buffer
)

type AppStatusRequest struct {
	AppUptime     int64         `json:"app_uptime"`
	Current       string        `json:"current"`
	IsOnline      *bool         `json:"is_online"`
	IsPluggedIn   *bool         `json:"is_plugged_in"`
	KernelUpdated *bool         `json:"kernel_updated"`
	LastHeard     time.Time     `json:"last_heard"`
	SystemUptime  time.Duration `json:"system_uptime"`
}

func getLastHeard() (time.Time, error) {
	f, err := os.Open(appLastHeardFilePath)
	if err != nil {
		_, ok := errors.AsType[*os.PathError](err)
		if !ok {
			return time.Time{}, err
		}
	}
	defer f.Close()

	appLastHeardBufMu.Lock()
	defer appLastHeardBufMu.Unlock()

	appLastHeardBuf.Reset()
	appLastHeardBuf.ReadFrom(f)

	if len(appLastHeardBuf.Bytes()) == 0 {
		return time.Time{}, fmt.Errorf("old last heard value missing/nil (getLastHeard)")
	}
	timeParsed, err := time.Parse(time.RFC3339, appLastHeardBuf.String())
	if err != nil {
		return time.Time{}, err
	}
	fmt.Printf("Old app last heard value: %s\n", timeParsed.Format(time.RFC3339))
	return timeParsed, nil
}

func updateLastHeard(t time.Time) error {
	if t.IsZero() {
		return fmt.Errorf("time value required (updateLastHeard)")
	}
	nowStrFormatted := t.Format(time.RFC3339)

	appLastHeardBufMu.Lock()
	defer appLastHeardBufMu.Unlock()

	appLastHeardBuf.Reset()
	appLastHeardBuf.WriteString(nowStrFormatted)

	if err := os.WriteFile(appLastHeardFilePath, appLastHeardBuf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("New app last heard value: %s\n", nowStrFormatted)
	return nil
}

func getAppUptime() (time.Duration, error) {
	f, err := os.Open(appUptimeFilePath)
	if err != nil {
		return 0, err
	}
	appUptimeBufMu.Lock()
	defer appUptimeBufMu.Unlock()

	appUptimeBuf.Reset()
	appUptimeBuf.ReadFrom(f)

	initialTimeVal, err := time.Parse(time.RFC3339, appUptimeBuf.String())
	if err != nil {
		return 0, fmt.Errorf("cannot parse initialTimeVal (getAppUptime): %w\n", err)
	}

	uptime := time.Since(initialTimeVal)

	fmt.Printf("New app uptime value: %s\n", uptime)
	return time.Duration(uptime), nil
}

func updateAppUptime(t time.Time) error {
	if t.IsZero() {
		return fmt.Errorf("time value required (updateAppUptime)")
	}
	appUptimeBufMu.Lock()
	defer appUptimeBufMu.Unlock()

	appUptimeBuf.Reset()
	appUptimeBuf.WriteString(time.Now().Format(time.RFC3339))

	if err := os.WriteFile(appUptimeFilePath, appUptimeBuf.Bytes(), 0644); err != nil {
		return err
	}
	fmt.Printf("Old app uptime value: %s\n", appUptimeBuf.String())
	return nil
}

func printAppData() {
	if _, err := getLastHeard(); err != nil {
		fmt.Printf("getLastHeard error: %v\n", err)
	}
	if err := updateLastHeard(time.Now()); err != nil {
		fmt.Printf("updateLastHeard error: %v\n", err)
	}
	if _, err := getAppUptime(); err != nil {
		fmt.Printf("getAppUptime error: %v\n", err)
	}
	if err := updateAppUptime(time.Now()); err != nil {
		fmt.Printf("updateAppUptime error: %v\n", err)
	}
}
