//go:build linux && amd64

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"
	"time"
)

type HTTPRequestPayload struct {
	RequestType     string  `json:"request_type"`
	Tagnumber       int64   `json:"tagnumber"`
	SystemSerial    string  `json:"system_serial"`
	Key             string  `json:"key"`
	StringValue     string  `json:"string_value,omitempty"`
	Value           any     `json:"value"`
	TransactionUUID *string `json:"transaction_uuid"`
}

const unixSocketPath = "/run/uit-client/uit-clientd.sock"

var (
	allowedKeys = []string{
		"battery_charge_pcnt",
		"bios_firmware",
		"bios_release_date",
		"bios_version",
		"chassis_type",
		"client_app_uptime",
		"client_lookup_by_serial",
		"cpu_core_count",
		"cpu_thread_count",
		"cpu_current_usage",
		"cpu_current_mhz",
		"cpu_manufacturer",
		"cpu_max_speed_mhz",
		"cpu_model",
		"cpu_millidegrees_c",
		"ethernet_mac",
		"init",
		"memory_capacity_kb",
		"memory_usage_kb",
		"motherboard_manufacturer",
		"motherboard_serial",
		"system_manufacturer",
		"system_model",
		"system_sku",
		"system_uptime",
		"system_uuid",
		"tpm_version",
		"wifi_mac",
	}
)

func isKeyAllowed(key string) bool {
	if slices.Contains(allowedKeys, key) {
		return true
	}
	return false
}

func expectedMethodForKey(key string) (string, bool) {
	switch key {
	case "client_lookup_by_serial":
		return "GET", true
	default:
		return "", false
	}
}

func isUUIDRequiredForKey(key string) bool {
	switch key {
	case "ethernet_mac":
		return true
	default:
		return false
	}
}

func main() {
	tagnumber := flag.Int64("tag", 0, "Tag number of client (optional, sent in addition to serial)")
	serial := flag.String("serial", "", "System serial number of client (required)")
	key := flag.String("key", "", "Key of request to send (required)")
	value := flag.String("value", "", "Value of request to send (required)")
	transactionUUID := flag.String("uuid", "", "Optional UUID of request/transaction")
	methodGET := flag.Bool("get", false, "Use GET method for the request (default is POST)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -serial <serial> -key <key> [-value <value>] [-tag <tagnumber>] [-uuid <uuid>] [-get]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "no arguments provided\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	httpPayload := new(HTTPRequestPayload)

	// Default POST if no method specified
	httpPayload.RequestType = "POST"
	if methodGET != nil {
		if *methodGET {
			httpPayload.RequestType = "GET"
		}
	}

	// Key is required
	if key == nil || *key == "" {
		fmt.Fprintf(os.Stderr, "key is required\n")
		os.Exit(1)
	}
	if !isKeyAllowed(*key) {
		fmt.Fprintf(os.Stderr, "key is not allowed\n")
		fmt.Fprintf(os.Stderr, "allowed keys are: %v\n", allowedKeys)
		os.Exit(1)
	}

	if expectedMethod, constrained := expectedMethodForKey(*key); constrained && httpPayload.RequestType != expectedMethod {
		fmt.Fprintf(os.Stderr, "key '%s' requires %s method\n", *key, expectedMethod)
		os.Exit(1)
	}

	httpPayload.Tagnumber = *tagnumber
	httpPayload.SystemSerial = *serial

	if serial == nil || strings.TrimSpace(*serial) == "" {
		fmt.Fprintf(os.Stderr, "serial is required\n")
		os.Exit(1)
	}

	if tagnumber != nil && *tagnumber != 0 && (*tagnumber < 1 || *tagnumber > 999999) {
		fmt.Fprintf(os.Stderr, "tag number must be between 1 and 999999 when provided\n")
		os.Exit(1)
	}

	if *key != "init" && *key != "client_lookup_by_serial" && (value == nil || strings.TrimSpace(*value) == "") {
		fmt.Fprintf(os.Stderr, "value is required\n")
		os.Exit(1)
	}
	if isUUIDRequiredForKey(*key) && (transactionUUID == nil || strings.TrimSpace(*transactionUUID) == "") {
		fmt.Fprintf(os.Stderr, "uuid is required for key '%s'\n", *key)
		os.Exit(1)
	}

	httpPayload.Key = *key

	// Value
	httpPayload.StringValue = *value
	if *key == "client_lookup_by_serial" && strings.TrimSpace(httpPayload.StringValue) == "" {
		httpPayload.StringValue = httpPayload.SystemSerial
	}

	// Transaction UUID
	if transactionUUID != nil && *transactionUUID != "" {
		httpPayload.TransactionUUID = transactionUUID
	} else {
		httpPayload.TransactionUUID = nil
	}

	conn, err := net.DialTimeout("unix", unixSocketPath, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s: %v\n", unixSocketPath, err)
		os.Exit(1)
	}
	defer conn.Close()

	line, err := json.Marshal(httpPayload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal payload: %v\n", err)
		os.Exit(1)
	}
	if _, err := io.WriteString(conn, string(line)+"\n"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		os.Exit(1)
	}

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set read deadline: %v\n", err)
		os.Exit(1)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read response: %v\n", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return
	}
	if strings.HasPrefix(response, "ERROR: ") {
		fmt.Fprintln(os.Stderr, strings.TrimPrefix(response, "ERROR: "))
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, response)
}
