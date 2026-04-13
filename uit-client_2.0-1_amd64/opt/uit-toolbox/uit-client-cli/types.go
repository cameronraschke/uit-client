//go:build linux && amd64

package main

import (
	"net/url"
)

type ClientConfig struct {
	UIT_CLIENT_DB_USER   string `json:"UIT_CLIENT_DB_USER"`
	UIT_CLIENT_DB_PASSWD string `json:"UIT_CLIENT_DB_PASSWD"`
	UIT_CLIENT_DB_NAME   string `json:"UIT_CLIENT_DB_NAME"`
	UIT_CLIENT_DB_HOST   string `json:"UIT_CLIENT_DB_HOST"`
	UIT_CLIENT_DB_PORT   string `json:"UIT_CLIENT_DB_PORT"`
	UIT_CLIENT_NTP_HOST  string `json:"UIT_CLIENT_NTP_HOST"`
	UIT_CLIENT_PING_HOST string `json:"UIT_CLIENT_PING_HOST"`
	UIT_SERVER_HOSTNAME  string `json:"UIT_SERVER_HOSTNAME"`
	UIT_WEB_HTTP_HOST    string `json:"UIT_WEB_HTTP_HOST"`
	UIT_WEB_HTTP_PORT    string `json:"UIT_WEB_HTTP_PORT"`
	UIT_WEB_HTTPS_HOST   string `json:"UIT_WEB_HTTPS_HOST"`
	UIT_WEB_HTTPS_PORT   string `json:"UIT_WEB_HTTPS_PORT"`
	UIT_WEBMASTER_NAME   string `json:"UIT_WEBMASTER_NAME"`
	UIT_WEBMASTER_EMAIL  string `json:"UIT_WEBMASTER_EMAIL"`
}

type HTTPRequest struct {
	Config  *HTTPRequestConfig
	Payload *HTTPRequestPayload
}

type HTTPRequestConfig struct {
	URL    url.URL
	Method string
}

type HTTPRequestPayload struct {
	Tagnumber int64
	Key       string
	Value     any
	UUID      *string
}

type CPUDataRequest struct {
	Tagnumber     *int64   `json:"tagnumber"`
	UsagePercent  *float64 `json:"cpu_current_usage"`
	MHz           *float64 `json:"cpu_current_mhz"`
	MillidegreesC *float64 `json:"cpu_millidegrees_c"`
}

type BatteryData struct {
	Tagnumber int64    `json:"tagnumber"`
	Percent   *float64 `json:"battery_charge_pcnt"`
}

type ClientUptime struct {
	Tagnumber       int64  `json:"tagnumber"`
	ClientAppUptime *int64 `json:"client_app_uptime"`
	SystemUptime    *int64 `json:"system_uptime"`
}

type MemoryDataRequest struct {
	Tagnumber       *int64  `json:"tagnumber"`
	TotalUsageKB    *int64  `json:"memory_usage_kb"`
	TotalCapacityKB *int64  `json:"memory_capacity_kb"`
	Type            *string `json:"type"`
	SpeedMHz        *int64  `json:"speed_mhz"`
}
