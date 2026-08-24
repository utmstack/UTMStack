package syslog

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Buffer size constants for syslog message processing.
const (
	MinBufferSize         = 480
	RecommendedBufferSize = 2048
	MaxBufferSize         = 8192
	UDPBufferSize         = 2048
)

// FramingMethod represents the syslog message framing method.
type FramingMethod int

const (
	FramingNewline FramingMethod = iota
	FramingOctetCounting
)

// detectFramingMethod determines whether the message uses octet counting or newline framing.
func detectFramingMethod(reader *bufio.Reader) (FramingMethod, error) {
	firstByte, err := reader.Peek(1)
	if err != nil {
		return 0, fmt.Errorf("failed to peek first byte: %w", err)
	}

	if firstByte[0] >= '0' && firstByte[0] <= '9' {
		return FramingOctetCounting, nil
	}

	if firstByte[0] == '<' {
		return FramingNewline, nil
	}

	return 0, fmt.Errorf("unknown framing method, first byte: 0x%02x", firstByte[0])
}

// readOctetCountingFrame reads a syslog message using octet counting framing (RFC 5425).
func readOctetCountingFrame(reader *bufio.Reader) (string, error) {
	lengthStr, err := reader.ReadString(' ')
	if err != nil {
		return "", fmt.Errorf("failed to read message length: %w", err)
	}

	lengthStr = strings.TrimSuffix(lengthStr, " ")
	msgLen, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("invalid message length '%s': %w", lengthStr, err)
	}

	if msgLen < 1 {
		return "", fmt.Errorf("message length %d is too small (minimum 1)", msgLen)
	}
	if msgLen > MaxBufferSize {
		return "", fmt.Errorf("message length %d exceeds maximum %d", msgLen, MaxBufferSize)
	}

	msgBytes := make([]byte, msgLen)
	_, err = io.ReadFull(reader, msgBytes)
	if err != nil {
		return "", fmt.Errorf("failed to read %d byte message body: %w", msgLen, err)
	}

	return string(msgBytes), nil
}

// readNewlineFrame reads a syslog message using newline-delimited framing.
func readNewlineFrame(reader *bufio.Reader) (string, error) {
	message, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read newline-delimited message: %w", err)
	}
	return message, nil
}

// readSyslogMessage reads a complete syslog message, auto-detecting the framing method.
func readSyslogMessage(reader *bufio.Reader) (string, error) {
	method, err := detectFramingMethod(reader)
	if err != nil {
		return "", err
	}

	switch method {
	case FramingOctetCounting:
		return readOctetCountingFrame(reader)
	case FramingNewline:
		return readNewlineFrame(reader)
	default:
		return "", fmt.Errorf("unsupported framing method: %d", method)
	}
}
