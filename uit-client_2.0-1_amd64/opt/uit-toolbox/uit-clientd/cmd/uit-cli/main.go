//go:build linux && amd64

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"time"
)

type HTTPRequestPayload struct {
	RequestType  string  `json:"request_type"`
	Tagnumber    int64   `json:"tagnumber"`
	SystemSerial string  `json:"system_serial"`
	Key          string  `json:"key"`
	StringValue  string  `json:"string_value,omitempty"`
	Value        any     `json:"value"`
	UUID         *string `json:"uuid,omitempty"`
}

const unixSocketPath = "/run/uit-client/uit-clientd.sock"

var (
	allowedKeys = []string{
		"battery_charge_pcnt",
		"client_app_uptime",
		"system_uptime",
		"memory_usage_kb",
		"memory_capacity_kb",
		"cpu_current_usage",
		"cpu_current_mhz",
		"cpu_millidegrees_c",
		"client_lookup_by_serial",
	}
)

func isKeyAllowed(key string) bool {
	if slices.Contains(allowedKeys, key) {
		return true
	}
	return false
}

func main() {
	tagnumber := flag.Int64("tag", 0, "Tag number of client (required)")
	serial := flag.String("serial", "", "System serial number of client (optional, used for lookups)")
	key := flag.String("key", "", "Key of request to send (required)")
	value := flag.String("value", "", "Value of request to send (required)")
	transactionUUID := flag.String("uuid", "", "Optional UUID of request/transaction")
	methodGET := flag.Bool("get", false, "Use GET method for the request (default is POST)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-tag <tagnumber> | -serial <serial>] -key <key> -value <value> [-uuid <uuid>] [-get]\n", os.Args[0])
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

	// No tag/serial needed for GET requests
	if (tagnumber == nil || *tagnumber == 0) && (serial == nil || *serial == "") {
		fmt.Fprintf(os.Stderr, "either tag number or serial number is required\n")
		os.Exit(1)
	}
	// No value needed for GET requests
	if methodGET != nil && *methodGET {
		if value == nil || *value == "" {
			fmt.Fprintf(os.Stderr, "value is required\n")
			os.Exit(1)
		}
	}
	httpPayload.Tagnumber = *tagnumber
	httpPayload.SystemSerial = *serial

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
	httpPayload.Key = *key

	// Value
	httpPayload.StringValue = *value

	// Transaction UUID
	httpPayload.UUID = transactionUUID

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
}
