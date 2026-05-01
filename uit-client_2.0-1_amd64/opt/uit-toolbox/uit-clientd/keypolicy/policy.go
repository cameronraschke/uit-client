//go:build linux && amd64

package keypolicy

import "sort"

type Policy struct {
	Method       string
	RequiresUUID bool
}

var policies = map[string]Policy{
	"battery_charge_pcnt":          {Method: "POST", RequiresUUID: false},
	"battery_charge_cycles":        {Method: "POST", RequiresUUID: true},
	"battery_current_max_capacity": {Method: "POST", RequiresUUID: true},
	"battery_design_capacity":      {Method: "POST", RequiresUUID: true},
	"battery_manufacture_date":     {Method: "POST", RequiresUUID: true},
	"battery_manufacturer":         {Method: "POST", RequiresUUID: true},
	"battery_model":                {Method: "POST", RequiresUUID: true},
	"battery_serial":               {Method: "POST", RequiresUUID: true},
	"bios_firmware":                {Method: "POST", RequiresUUID: true},
	"bios_release_date":            {Method: "POST", RequiresUUID: true},
	"bios_version":                 {Method: "POST", RequiresUUID: true},
	"chassis_type":                 {Method: "POST", RequiresUUID: true},
	"client_app_uptime":            {Method: "POST", RequiresUUID: false},
	"client_lookup_by_serial":      {Method: "GET", RequiresUUID: false},
	"cpu_core_count":               {Method: "POST", RequiresUUID: true},
	"cpu_thread_count":             {Method: "POST", RequiresUUID: true},
	"cpu_current_usage":            {Method: "POST", RequiresUUID: false},
	"cpu_current_mhz":              {Method: "POST", RequiresUUID: false},
	"cpu_manufacturer":             {Method: "POST", RequiresUUID: true},
	"cpu_max_speed_mhz":            {Method: "POST", RequiresUUID: true},
	"cpu_model":                    {Method: "POST", RequiresUUID: true},
	"cpu_millidegrees_c":           {Method: "POST", RequiresUUID: false},
	"disk_errors":                  {Method: "POST", RequiresUUID: true},
	"disk_firmware":                {Method: "POST", RequiresUUID: true},
	"disk_model":                   {Method: "POST", RequiresUUID: true},
	"disk_power_cycles":            {Method: "POST", RequiresUUID: true},
	"disk_power_on_hours":          {Method: "POST", RequiresUUID: true},
	"disk_reads_kb":                {Method: "POST", RequiresUUID: true},
	"disk_serial":                  {Method: "POST", RequiresUUID: true},
	"disk_size_kb":                 {Method: "POST", RequiresUUID: true},
	"disk_type":                    {Method: "POST", RequiresUUID: true},
	"disk_writes_kb":               {Method: "POST", RequiresUUID: true},
	"ethernet_mac":                 {Method: "POST", RequiresUUID: true},
	"init":                         {Method: "POST", RequiresUUID: true},
	"memory_capacity_kb":           {Method: "POST", RequiresUUID: true},
	"memory_serial":                {Method: "POST", RequiresUUID: true},
	"memory_speed_mhz":             {Method: "POST", RequiresUUID: true},
	"memory_usage_kb":              {Method: "POST", RequiresUUID: false},
	"motherboard_manufacturer":     {Method: "POST", RequiresUUID: true},
	"motherboard_serial":           {Method: "POST", RequiresUUID: true},
	"system_manufacturer":          {Method: "POST", RequiresUUID: true},
	"system_model":                 {Method: "POST", RequiresUUID: true},
	"system_sku":                   {Method: "POST", RequiresUUID: true},
	"system_uptime":                {Method: "POST", RequiresUUID: false},
	"system_uuid":                  {Method: "POST", RequiresUUID: true},
	"tpm_version":                  {Method: "POST", RequiresUUID: true},
	"wifi_mac":                     {Method: "POST", RequiresUUID: true},
}

func Lookup(key string) (Policy, bool) {
	policy, ok := policies[key]
	return policy, ok
}

func AllowedKeys() []string {
	keys := make([]string, 0, len(policies))
	for key := range policies {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
