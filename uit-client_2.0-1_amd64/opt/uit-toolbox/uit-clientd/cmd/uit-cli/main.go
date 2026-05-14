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
	"strings"
	"time"

	"uit-clientd/keypolicy"
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

func main() {
	tagnumber := flag.Int64("tag", 0, "Tag number of client (optional, sent in addition to serial)")
	serial := flag.String("serial", "", "System serial number of client (required)")
	key := flag.String("key", "", "Key of request to send (required)")
	value := flag.String("value", "", "Value of request to send (required)")
	transactionUUID := flag.String("uuid", "", "Optional UUID of request/transaction")
	methodGET := flag.Bool("get", false, "Use GET method for the request (default is POST)")
	methodPOST := flag.Bool("post", false, "Use POST method for the request (default is POST)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -serial <serial> -key <key> [-value <value>] [-tag <tagnumber>] [-uuid <uuid>] [-get] [-post]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "no arguments provided\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	httpPayload := new(HTTPRequestPayload)

	// common logical errors
	if *methodGET && *methodPOST {
		fmt.Fprintf(os.Stderr, "cannot specify both -get and -post\n")
		os.Exit(1)
	}

	// Key is required
	if key == nil || *key == "" {
		fmt.Fprintf(os.Stderr, "key is required\n")
		os.Exit(1)
	}
	rule, ok := keypolicy.Lookup(*key)
	if !ok {
		fmt.Fprintf(os.Stderr, "key is not allowed\n")
		fmt.Fprintf(os.Stderr, "allowed keys are: %v\n", keypolicy.AllowedKeys())
		os.Exit(1)
	}
	httpPayload.Key = *key

	// HTTP method check
	if *methodGET {
		httpPayload.RequestType = "GET"
	} else if *methodPOST {
		httpPayload.RequestType = "POST"
	}
	if rule.Method != "" && httpPayload.RequestType != rule.Method {
		fmt.Fprintf(os.Stderr, "key '%s' requires %s method\n", *key, rule.Method)
		os.Exit(1)
	}
	httpPayload.RequestType = rule.Method

	// Tagnumber checks
	if rule.RequiresTag && *tagnumber == 0 {
		fmt.Fprintf(os.Stderr, "tag number is required for key '%s'\n", *key)
		os.Exit(1)
	}
	if *tagnumber != 0 && (*tagnumber < 100000 || *tagnumber > 999999) {
		fmt.Fprintf(os.Stderr, "tag number must be between 100000 and 999999\n")
		os.Exit(1)
	}
	httpPayload.Tagnumber = *tagnumber

	// Serial checks
	if rule.RequiresSerial && (serial == nil || strings.TrimSpace(*serial) == "") {
		fmt.Fprintf(os.Stderr, "serial is required for key '%s'\n", *key)
		os.Exit(1)
	}
	httpPayload.SystemSerial = *serial

	// UUID checks
	if rule.RequiresUUID && (transactionUUID == nil || strings.TrimSpace(*transactionUUID) == "") {
		fmt.Fprintf(os.Stderr, "uuid is required for key '%s'\n", *key)
		os.Exit(1)
	}
	// Transaction UUID
	if transactionUUID != nil && *transactionUUID != "" {
		httpPayload.TransactionUUID = transactionUUID
	} else {
		httpPayload.TransactionUUID = nil
	}

	// Value
	// exceptions for certain keys
	if rule.RequiresValue && (value == nil || strings.TrimSpace(*value) == "") {
		fmt.Fprintf(os.Stderr, "value is required for key '%s'\n", *key)
		os.Exit(1)
	}
	httpPayload.StringValue = *value
	if *key == "client_lookup_by_serial" && strings.TrimSpace(httpPayload.StringValue) == "" {
		httpPayload.StringValue = httpPayload.SystemSerial
	}

	// connect to unix socket
	conn, err := net.DialTimeout("unix", unixSocketPath, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s: %v\n", unixSocketPath, err)
		os.Exit(1)
	}
	defer conn.Close()

	// encode json & send line
	line, err := json.Marshal(httpPayload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal payload: %v\n", err)
		os.Exit(1)
	}
	if _, err := io.WriteString(conn, string(line)+"\n"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		os.Exit(1)
	}
	if httpPayload.RequestType != "GET" {
		return
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
