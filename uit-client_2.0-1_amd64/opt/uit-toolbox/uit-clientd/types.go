//go:build linux && amd64

package main

import (
	"net/url"
)

type ParsedInputDTO struct {
	RequestType string  `json:"request_type"`
	Tagnumber   int64   `json:"tagnumber"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	UUID        *string `json:"uuid,omitempty"`
}

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
	RequestType     string  `json:"request_type"`
	Tagnumber       int64   `json:"tagnumber"`
	SystemSerial    string  `json:"system_serial"`
	Key             string  `json:"key"`
	StringValue     string  `json:"string_value,omitempty"`
	Value           any     `json:"value"`
	TransactionUUID *string `json:"transaction_uuid,omitempty"`
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

type ClientInitRequest struct {
	Tagnumber       *int64  `json:"tagnumber"`
	SystemSerial    *string `json:"system_serial"`
	TransactionUUID *string `json:"transaction_uuid,omitempty"`
}

type ClientHardwareView struct {
	TransactionUUID           string   `json:"transaction_uuid,omitempty"`
	Tagnumber                 *int64   `json:"tagnumber,omitempty"`
	SystemSerial              *string  `json:"system_serial,omitempty"`
	SystemUUID                *string  `json:"system_uuid,omitempty"`
	SystemManufacturer        *string  `json:"system_manufacturer,omitempty"`
	SystemModel               *string  `json:"system_model,omitempty"`
	SystemSKU                 *string  `json:"system_sku,omitempty"`
	ProductFamily             *string  `json:"product_family,omitempty"`
	ProductName               *string  `json:"product_name,omitempty"`
	DeviceType                *string  `json:"device_type,omitempty"`
	ChassisType               *string  `json:"chassis_type,omitempty"`
	MotherboardSerial         *string  `json:"motherboard_serial,omitempty"`
	MotherboardManufacturer   *string  `json:"motherboard_manufacturer,omitempty"`
	CPUManufacturer           *string  `json:"cpu_manufacturer,omitempty"`
	CPUModel                  *string  `json:"cpu_model,omitempty"`
	CPUMaxSpeedMhz            *int64   `json:"cpu_max_speed_mhz,omitempty"`
	CPUCoreCount              *int64   `json:"cpu_core_count,omitempty"`
	CPUThreadCount            *int64   `json:"cpu_thread_count,omitempty"`
	EthernetMAC               *string  `json:"ethernet_mac,omitempty"`
	WiFiMAC                   *string  `json:"wifi_mac,omitempty"`
	TPMVersion                *string  `json:"tpm_version,omitempty"`
	DiskModel                 *string  `json:"disk_model,omitempty"`
	DiskType                  *string  `json:"disk_type,omitempty"`
	DiskSize                  *int64   `json:"disk_size_kb,omitempty"`
	DiskSerial                *string  `json:"disk_serial,omitempty"`
	DiskWritesKB              *int64   `json:"disk_writes_kb,omitempty"`
	DiskReadsKB               *int64   `json:"disk_reads_kb,omitempty"`
	DiskPowerOnHours          *int64   `json:"disk_power_on_hours,omitempty"`
	DiskErrors                *int64   `json:"disk_errors,omitempty"`
	DiskPowerCycles           *int64   `json:"disk_power_cycles,omitempty"`
	DiskFirmware              *string  `json:"disk_firmware,omitempty"`
	BatteryModel              *string  `json:"battery_model,omitempty"`
	BatterySerial             *string  `json:"battery_serial,omitempty"`
	BatteryChargeCycles       *int64   `json:"battery_charge_cycles,omitempty"`
	BatteryCurrentMaxCapacity *float64 `json:"battery_current_max_capacity,omitempty"`
	BatteryDesignCapacity     *float64 `json:"battery_design_capacity,omitempty"`
	BatteryManufacturer       *string  `json:"battery_manufacturer,omitempty"`
	BatteryManufactureDate    *string  `json:"battery_manufacture_date,omitempty"`
	BiosVersion               *string  `json:"bios_version,omitempty"`
	BiosReleaseDate           *string  `json:"bios_release_date,omitempty"`
	BiosFirmware              *string  `json:"bios_firmware,omitempty"`
	MemorySerial              *string  `json:"memory_serial,omitempty"`
	MemoryCapacityKB          *int64   `json:"memory_capacity_kb,omitempty"`
	MemorySpeedMHz            *int64   `json:"memory_speed_mhz,omitempty"`
}
