//go:build linux && amd64

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

func getUnixSocketListener() (net.Listener, bool, error) {
	listener, err := getInheritedUnixSocketListener()
	if err == nil {
		return listener, true, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}

	listener, err = net.Listen("unix", unixSocketPath)
	if err != nil {
		return nil, false, err
	}

	if err := os.Chmod(unixSocketPath, 0660); err != nil {
		_ = listener.Close()
		_ = os.Remove(unixSocketPath)
		return nil, false, fmt.Errorf("failed to chmod unix socket: %w", err)
	}

	return listener, false, nil
}

func getInheritedUnixSocketListener() (net.Listener, error) {
	listenPID := os.Getenv("LISTEN_PID")
	listenFDs := os.Getenv("LISTEN_FDS")
	if listenPID == "" || listenFDs == "" {
		return nil, os.ErrNotExist
	}

	pid, err := strconv.Atoi(listenPID)
	if err != nil {
		return nil, fmt.Errorf("invalid LISTEN_PID: %w", err)
	}
	if pid != os.Getpid() { // LISTEN_PID must match the current PID
		return nil, os.ErrNotExist
	}

	fds, err := strconv.Atoi(listenFDs)
	if err != nil {
		return nil, fmt.Errorf("invalid LISTEN_FDS: %w", err)
	}
	if fds < 1 {
		return nil, os.ErrNotExist
	}

	file := os.NewFile(uintptr(3), "systemd-unix-socket") // first systemd fd is # 3
	if file == nil {
		return nil, fmt.Errorf("failed to access inherited systemd socket")
	}
	defer file.Close()

	listener, err := net.FileListener(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener from inherited socket: %w", err)
	}

	return listener, nil
}

func handleConnection(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		response, err := handleInput(ctx, scanner.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to handle input: %v\n", err)
			_, _ = fmt.Fprintf(conn, "ERROR: %v\n", err)
			continue
		}

		_, _ = fmt.Fprintf(conn, "%s\n", response)
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "unix socket read error: %v\n", err)
	}
}
