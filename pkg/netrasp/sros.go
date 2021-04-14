package netrasp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// SROS is the Netrasp driver for Nokia SR OS (MD-CLI).
type sros struct {
	Connection connection
}

// Close connection to device.
func (i sros) Close(ctx context.Context) error {
	i.Connection.Close(ctx)

	return nil
}

// Configure device.
func (i sros) Configure(ctx context.Context, commands []string) (string, error) {
	var output string
	_, err := i.Run(ctx, "edit-config exclusive")
	if err != nil {
		return "", fmt.Errorf("unable to enter exclusive edit mode: %w", err)
	}
	for _, command := range commands {
		result, err := i.Run(ctx, command)
		if err != nil {
			return output, fmt.Errorf("unable to run command '%s': %w", command, err)
		}
		output += result
	}
	_, err = i.Run(ctx, "commit")
	if err != nil {
		return output, fmt.Errorf("unable to commit configuration: %w", err)
	}

	return output, nil
}

// Dial opens a connection to a device.
func (i sros) Dial(ctx context.Context) error {
	return establishConnection(ctx, i, i.Connection, i.basePrompt(), "environment more false")
}

// Enable elevates privileges.
func (i sros) Enable(ctx context.Context) error {
	return nil
}

// Run executes a command on a device.
func (i sros) Run(ctx context.Context, command string) (string, error) {
	output, err := i.RunUntil(ctx, command, i.basePrompt())
	if err != nil {
		return "", err
	}

	output = strings.ReplaceAll(output, "\r\n", "\n")
	lines := strings.Split(output, "\n")
	result := ""
	// len-2 to cut off the context piece of the prompt aka [/]
	for i := 1; i < len(lines)-2; i++ {
		// skip empty lines that sros adds for visual separation of diff commands
		// as we add it manually
		if (i == len(lines)-3) && (lines[i] == "") {
			continue
		} else {
			result += lines[i] + "\n"
		}

	}

	return result, nil
}

// RunUntil executes a command and reads until the provided prompt.
func (i sros) RunUntil(ctx context.Context, command string, prompt *regexp.Regexp) (string, error) {
	err := i.Connection.Send(ctx, command)
	if err != nil {
		return "", fmt.Errorf("unable to send command to device: %w", err)
	}

	reader := i.Connection.Recv(ctx)
	output, err := readUntilPrompt(ctx, reader, prompt)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (i sros) basePrompt() *regexp.Regexp {
	return regexp.MustCompile(`^[ABCD]:\S+@\S+#`)
}
