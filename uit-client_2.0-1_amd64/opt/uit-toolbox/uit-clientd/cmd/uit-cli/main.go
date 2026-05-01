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
	TransactionUUID *string `json:"transaction_uuid,omitempty"`
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
		"init",
		"ethernet_mac",
		"wifi_mac",
		"tpm_version",
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

	if httpPayload.RequestType == "POST" {
		if tagnumber == nil || *tagnumber <= 0 || *tagnumber > 999999 {
			fmt.Fprintf(os.Stderr, "tag number is required and must be between 1 and 999999 for POST requests\n")
			os.Exit(1)
		}
	} else if tagnumber != nil && *tagnumber != 0 && (*tagnumber < 0 || *tagnumber > 999999) {
		fmt.Fprintf(os.Stderr, "tag number must be between 1 and 999999 when provided\n")
		os.Exit(1)
	}

	if *key != "init" && (value == nil || *value == "") {
		fmt.Fprintf(os.Stderr, "value is required\n")
		os.Exit(1)
	}

	httpPayload.Key = *key

	// Value
	httpPayload.StringValue = *value

	// Transaction UUID
	httpPayload.TransactionUUID = transactionUUID

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
