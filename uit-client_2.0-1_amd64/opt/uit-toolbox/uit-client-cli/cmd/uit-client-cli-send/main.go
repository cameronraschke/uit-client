package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const unixSocketPath = "/run/uit-client/uit-client-cli.sock"

func main() {
	line, err := inputLine(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	conn, err := net.DialTimeout("unix", unixSocketPath, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to %s: %v\n", unixSocketPath, err)
		os.Exit(1)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, line+"\n"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		os.Exit(1)
	}
}

func inputLine(args []string) (string, error) {
	if len(args) > 0 {
		line := strings.TrimSpace(strings.Join(args, " "))
		if line == "" {
			return "", fmt.Errorf("input line cannot be empty")
		}
		return line, nil
	}

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	line := strings.TrimSpace(string(stdin))
	if line == "" {
		return "", fmt.Errorf("usage: uit-client-cli-send 'TAG|KEY|VALUE[|UUID]' or pipe a line on stdin")
	}

	return line, nil
}
