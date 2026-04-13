//go:build linux && amd64

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"time"
)

const unixSocketPath = "/run/uit-client/uit-client-cli.sock"

var (
	allowedKeys = []string{
		"battery_charge_pcnt",
		"client_app_uptime",
		"system_uptime",
		"cpu_usage",
		"memory_usage_kb",
		"memory_capacity_kb",
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
	key := flag.String("key", "", "Key of request to send (required)")
	value := flag.String("value", "", "Value of request to send (required)")
	uuid := flag.String("uuid", "", "Optional UUID of request/transaction")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -tag <tagnumber> -key <key> -value <value> [-uuid <uuid>]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "no arguments provided\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if tagnumber == nil || *tagnumber == 0 {
		fmt.Fprintf(os.Stderr, "tag number is required\n")
		os.Exit(1)
	}
	if key == nil || *key == "" {
		fmt.Fprintf(os.Stderr, "key is required\n")
		os.Exit(1)
	}
	if value == nil || *value == "" {
		fmt.Fprintf(os.Stderr, "value is required\n")
		os.Exit(1)
	}
	if !isKeyAllowed(*key) {
		fmt.Fprintf(os.Stderr, "key is not allowed\n")
		fmt.Fprintf(os.Stderr, "allowed keys are: %v\n", allowedKeys)
		os.Exit(1)
	}

	conn, err := net.DialTimeout("unix", unixSocketPath, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s: %v\n", unixSocketPath, err)
		os.Exit(1)
	}
	defer conn.Close()

	line := fmt.Sprintf("%d|%s|%s", *tagnumber, *key, *value)
	if *uuid != "" {
		line = fmt.Sprintf("%d|%s|%s|%s", *tagnumber, *key, *value, *uuid)
	}
	if _, err := io.WriteString(conn, line+"\n"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		os.Exit(1)
	}
}
