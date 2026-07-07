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

func getUnixSocketConnection() (net.Conn, error) {
	conn, err := net.DialTimeout("unix", unixSocketPath, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return nil, err
	}
	return conn, nil
}

func sendDataToSocket(conn net.Conn, payload HTTPRequestPayload) error {
	// encode json & send line
	line, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}
	if _, err := io.WriteString(conn, string(line)+"\n"); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func readResponseFromSocket(conn net.Conn) (string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return "", fmt.Errorf("failed to set read deadline: %v", err)
	}
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return "", nil
	}
	if strings.HasPrefix(response, "ERROR: ") {
		return "", fmt.Errorf("%s", strings.TrimPrefix(response, "ERROR: "))
	}
	return response, nil
}

func main() {
	serial := flag.String("serial", "", "System serial number of client (required)")
	tagnumber := flag.Int64("tag", 0, "Tag number of client (optional)")
	key := flag.String("key", "", "Key of request to send")
	value := flag.String("value", "", "Value associated with the key")
	transactionUUID := flag.String("uuid", "", "Request/transaction UUID (optional)")
	methodGET := flag.Bool("get", false, "Use GET method for the request (default is POST)")
	methodPOST := flag.Bool("post", false, "Use POST method for the request (default is POST)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cli: Usage: %s --serial <serial> [--tag <tagnumber>] --key <key> [--value <value>] [--uuid <uuid>] [--get | --post]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "cli: no arguments provided\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	httpPayload := new(HTTPRequestPayload)

	// common logical errors
	// cannot specify both GET and POST
	if *methodGET && *methodPOST {
		fmt.Fprintf(os.Stderr, "cli: cannot specify both --get and --post\n")
		os.Exit(1)
	}

	// Check key policy rules
	// Key is required before lookup
	if key == nil || *key == "" {
		fmt.Fprintf(os.Stderr, "cli: key is required\n")
		os.Exit(1)
	}
	rule, ok := keypolicy.Lookup(*key)
	if !ok {
		fmt.Fprintf(os.Stderr, "cli: key '%s' is not allowed\n", *key)
		fmt.Fprintf(os.Stderr, "allowed keys are: %v\n", keypolicy.AllowedKeys())
		os.Exit(1)
	}
	httpPayload.Key = *key

	// Serial checks
	if rule.RequiresSerial && (serial == nil || strings.TrimSpace(*serial) == "") {
		fmt.Fprintf(os.Stderr, "cli: serial is required for key '%s'\n", *key)
		os.Exit(1)
	}
	httpPayload.SystemSerial = *serial

	// Tagnumber checks
	if rule.RequiresTag && *tagnumber == 0 {
		fmt.Fprintf(os.Stderr, "cli: tag number is required for key '%s'\n", *key)
		os.Exit(1)
	}
	if *tagnumber != 0 && (*tagnumber < 100000 || *tagnumber > 999999) {
		fmt.Fprintf(os.Stderr, "cli: tag number must be between 100000 and 999999 for key '%s'\n", *key)
		os.Exit(1)
	}
	httpPayload.Tagnumber = *tagnumber

	// Value
	// exceptions for certain keys
	if rule.RequiresValue && (value == nil || strings.TrimSpace(*value) == "") {
		fmt.Fprintf(os.Stderr, "cli: value is required for key '%s'\n", *key)
		os.Exit(1)
	}
	httpPayload.StringValue = *value
	if *key == "client_lookup_by_serial" && strings.TrimSpace(httpPayload.StringValue) == "" {
		// For client_lookup_by_serial, the value is the serial number
		httpPayload.StringValue = httpPayload.SystemSerial
	}

	// UUID checks
	if rule.RequiresUUID && (transactionUUID == nil || strings.TrimSpace(*transactionUUID) == "") {
		fmt.Fprintf(os.Stderr, "cli: uuid is required for key '%s'\n", *key)
		os.Exit(1)
	}
	// Transaction UUID
	if transactionUUID != nil && *transactionUUID != "" {
		httpPayload.TransactionUUID = transactionUUID
	} else {
		httpPayload.TransactionUUID = nil
	}

	// HTTP method check
	if *methodGET {
		httpPayload.RequestType = "GET"
	} else if *methodPOST {
		httpPayload.RequestType = "POST"
	} else {
		// default to POST if not specified
		httpPayload.RequestType = "POST"
	}
	if rule.Method != "" && httpPayload.RequestType != rule.Method {
		fmt.Fprintf(os.Stderr, "cli: key '%s' requires %s method\n", *key, rule.Method)
		os.Exit(1)
	}
	httpPayload.RequestType = rule.Method

	// connect to unix socket
	conn, err := getUnixSocketConnection()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cli: failed to connect to %s: %v\n", unixSocketPath, err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := sendDataToSocket(conn, *httpPayload); err != nil {
		fmt.Fprintf(os.Stderr, "cli: failed to write to socket: %v\n", err)
		os.Exit(1)
	}

	response, err := readResponseFromSocket(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cli: failed to read response from socket: %v\n", err)
		os.Exit(1)
	}

	if response == "" {
		fmt.Fprintf(os.Stdout, "cli: no response received from uit-clientd\n")
		os.Exit(0)
	}

	fmt.Fprintf(os.Stdout, "%s\n", response)
}
