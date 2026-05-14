//go:build linux && amd64

package keypolicy

import "sort"

type Policy struct {
	Method         string
	RequiresSerial bool
	RequiresTag    bool
	RequiresUUID   bool
	RequiresValue  bool
}

var policies = map[string]Policy{
	"battery_charge_pcnt":          {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"battery_charge_cycles":        {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"battery_current_max_capacity": {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"battery_design_capacity":      {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"battery_manufacture_date":     {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"battery_manufacturer":         {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"battery_model":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"battery_serial":               {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"bios_firmware":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"bios_release_date":            {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"bios_version":                 {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"chassis_type":                 {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"client_app_uptime":            {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"client_lookup_by_serial":      {Method: "GET", RequiresSerial: true, RequiresTag: false, RequiresUUID: false, RequiresValue: false},
	"clone_completed":              {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"clone_image_name":             {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"clone_job_duration":           {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"clone_master":                 {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"cpu_core_count":               {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"cpu_thread_count":             {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"cpu_current_usage":            {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"cpu_current_mhz":              {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"cpu_manufacturer":             {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"cpu_max_speed_mhz":            {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"cpu_model":                    {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"cpu_millidegrees_c":           {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"disk_errors":                  {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_firmware":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_model":                   {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_name":                    {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_power_cycles":            {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_power_on_hours":          {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_reads_kb":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_serial":                  {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_size_kb":                 {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_type":                    {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"disk_writes_kb":               {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"erase_completed":              {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"erase_disk_pcnt":              {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"erase_job_duration":           {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"erase_mode":                   {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"ethernet_mac":                 {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"init":                         {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: false},
	"job_cancelled":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"job_start_time":               {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"memory_capacity_kb":           {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"memory_serial":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"memory_speed_mhz":             {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"memory_usage_kb":              {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"motherboard_manufacturer":     {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"motherboard_serial":           {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"new_transaction_uuid":         {Method: "GET", RequiresSerial: false, RequiresTag: false, RequiresUUID: false, RequiresValue: false},
	"system_manufacturer":          {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"system_model":                 {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"system_sku":                   {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"system_uptime":                {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: false, RequiresValue: true},
	"system_uuid":                  {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"tpm_version":                  {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
	"wifi_mac":                     {Method: "POST", RequiresSerial: true, RequiresTag: true, RequiresUUID: true, RequiresValue: true},
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
