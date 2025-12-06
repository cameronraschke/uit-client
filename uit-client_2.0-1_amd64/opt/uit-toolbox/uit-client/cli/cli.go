//go:build linux && amd64

package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type CLI struct {
	T      *term.Terminal
	Prompt string
}
type Handler func(*CLI, []string) error

const (
	termWidth  = 80
	termHeight = 33
)

var (
	termFd     = int(os.Stdin.Fd())
	menuReader = bufio.NewReader(os.Stdin)
	menuWriter = bufio.NewWriter(os.Stdout)
)

var commands = map[string]Handler{
	"help": cmdHelp,
	"list": cmdList,
	"get":  cmdGet,
	"set":  cmdSet,
	"demo": cmdDemo,
	"exit": cmdExit,
}

func (cli *CLI) Run() error {
	for {
		cli.T.SetPrompt(cli.Prompt)
		line, err := cli.T.ReadLine()
		if err != nil {
			return err // Ctrl-D (EOF) exits
		}
		str := strings.TrimSpace(line)
		if str == "" {
			continue
		}
		parts := strings.Fields(str)
		cmd, args := parts[0], parts[1:]
		handler, ok := commands[cmd]
		if !ok {
			cli.Echo("Unknown command: %s", cmd)
			continue
		}
		if err := handler(cli, args); err != nil {
			cli.Echo("Error: %v", err)
		}
	}
}

func cmdHelp(cli *CLI, _ []string) error {
	cli.Echo("Commands: help, list, get <key>, set <key> <val>, demo, exit")
	return nil
}
func cmdExit(cli *CLI, _ []string) error {
	cli.Echo("Exiting...")
	return io.EOF
}

// caller
func cmdList(c *CLI, _ []string) error {
	c.Echo("Listing items:")
	c.Echo("Item 1")
	c.Echo("Item 2")
	c.Echo("Item 3")
	return nil
}
func cmdGet(cli *CLI, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: get <key>")
	}

	key := args[0]
	key = strings.Trim(key, "'\"")
	cli.Echo("%s = '%s'", key, "value")
	return nil
}
func cmdSet(cli *CLI, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: set <key> <value>")
	}
	key := args[0]
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	val := strings.Join(args[1:], " ")
	val = stripOuterQuotes(val)
	if val == "" {
		return fmt.Errorf("value cannot be empty")
	}
	cli.Echo("set %s to '%s'", key, val)
	return nil
}

func stripOuterQuotes(s string) string {
	if s == "" {
		return s
	}
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return s
	}
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")
	return s
}

// cmdDemo prints example output, reads input, then echoes it
func cmdDemo(cli *CLI, _ []string) error {
	cli.Echo("Example output: Welcome to the demo!")
	input, err := cli.Read("Type something: ")
	if err != nil {
		cli.Echo("Read error: %v", err)
		return err
	}
	cli.Echo("You typed: %s", input)
	return nil
}

// InitTerminalWithRaw controls raw mode; when raw=false, the OS echoes input
func InitTerminalWithRaw(raw bool) error {
	if !term.IsTerminal(termFd) {
		fmt.Printf("Warning: Standard input is not a terminal\n")
		return fmt.Errorf("standard input is not a terminal")
	}
	var old *term.State
	var err error
	if raw {
		old, err = term.MakeRaw(termFd)
		if err != nil {
			return err
		}
		// Ensure restore from the caller using the returned function
		defer term.Restore(termFd, old)
	}

	// Use unbuffered ReadWriter to avoid prompt delays
	type stdioRW struct {
		io.Reader
		io.Writer
	}
	rw := stdioRW{Reader: os.Stdin, Writer: os.Stdout}
	t := term.NewTerminal(rw, "> ")
	cli := &CLI{T: t, Prompt: "> "}

	// Handle Ctrl-C gracefully: do not exit; just print a note
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT)
	go func() {
		for range sigCh {
			// Mimic bash: on Ctrl-C, print a newline and re-show prompt
			fmt.Fprint(cli.T, "\n")
			menuWriter.Flush()
			cli.T.SetPrompt(cli.Prompt)
		}
	}()

	// Ensure any buffered writer state starts clean (kept for safety)
	menuWriter.Flush()

	// Optional size warning
	if curWidth, curHeight, err := term.GetSize(termFd); err == nil {
		if curWidth < termWidth || curHeight < termHeight {
			cli.Echo("Warning: Terminal smaller than recommended (%dx%d), current %dx%d", termWidth, termHeight, curWidth, curHeight)
		}
	}

	if err := cli.Run(); err != nil && err != io.EOF {
		cli.Echo("Exited: %v", err)
	}
	return nil
}

func (cli *CLI) ReadString(prompt string) (string, error) {
	s, err := cli.Read(prompt)
	return s, err
}
func (cli *CLI) ReadInt(prompt string) (int, error) {
	s, _ := cli.ReadString(prompt)
	return strconv.Atoi(s)
}
func (cli *CLI) ReadChoice(prompt string, choices []string) (string, error) {
	s, err := cli.Read(prompt)
	if err != nil {
		return "", err
	}
	for _, c := range choices {
		if s == c {
			return s, nil
		}
	}
	return "", fmt.Errorf("invalid choice: %s", s)
}

func (cli *CLI) Echo(format string, a ...any) {
	if format == "" {
		format = "%s"
	}
	fmt.Fprintf(cli.T, format+"\n", a...)
	menuWriter.Flush()
}

func (cli *CLI) Read(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(cli.T, prompt)
		menuWriter.Flush()
		cli.T.SetPrompt("")
	} else {
		cli.T.SetPrompt(cli.Prompt)
	}
	line, err := cli.T.ReadLine()
	if err != nil {
		return "", err
	}
	cli.T.SetPrompt("> ")
	input := strings.TrimSpace(line)
	if input == "" {
		return "", fmt.Errorf("no input received")
	}
	return input, nil
}

func (cli *CLI) ReadOneChar(prompt string) (rune, error) {
	runeBuf := make([]rune, 1)
	input, err := cli.Read(prompt)
	if err != nil {
		cli.Echo("Error reading input: %v", err)
		return 0, err
	}
	runeBuf = []rune(input)
	if len(runeBuf) == 0 {
		return 0, fmt.Errorf("no input received")
	}
	return runeBuf[0], nil
}

func RestoreTerminal(oldState *term.State) {
	if oldState != nil {
		_ = term.Restore(termFd, oldState)
	}
}
