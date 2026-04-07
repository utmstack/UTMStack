package agent

import (
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamAction indicates what the caller should do after handling a gRPC error.
type StreamAction int

const (
	// ActionContinue means retry the operation (non-fatal error).
	ActionContinue StreamAction = iota
	// ActionReconnect means break the inner loop and reconnect the stream.
	ActionReconnect
)

// HandleGRPCStreamError processes a gRPC stream error and returns the appropriate action.
// It handles EOF, Unavailable, and Canceled errors with deduplication of log messages.
// The errorLogged pointer tracks whether an error has already been logged to avoid spam.
func HandleGRPCStreamError(err error, msg string, errorLogged *bool) StreamAction {
	// EOF means the stream was closed by the server - reconnect
	if strings.Contains(err.Error(), "EOF") {
		utils.Logger.LogF(100, "%s: %v", msg, err)
		time.Sleep(timeToSleep)
		return ActionReconnect
	}

	st, ok := status.FromError(err)
	isTransient := ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled)

	// Log error only once to avoid spam
	logError(err, msg, errorLogged)
	time.Sleep(timeToSleep)

	if isTransient {
		// Transient errors (Unavailable, Canceled) require reconnection
		return ActionReconnect
	}

	// Other errors - retry the operation
	return ActionContinue
}

// logError logs an error message with deduplication.
// After the first error, subsequent errors are logged at debug level.
func logError(err error, msg string, errorLogged *bool) {
	if !*errorLogged {
		utils.Logger.ErrorF("%s: %v", msg, err)
		*errorLogged = true
	} else {
		utils.Logger.LogF(100, "%s: %v", msg, err)
	}
}

// LogConnectionError logs a connection error with deduplication.
func LogConnectionError(err error, target string, errorLogged *bool) {
	if !*errorLogged {
		utils.Logger.ErrorF("error connecting to %s: %v", target, err)
		*errorLogged = true
	} else {
		utils.Logger.LogF(100, "error connecting to %s: %v", target, err)
	}
}

// LogStreamError logs a stream creation error with deduplication.
func LogStreamError(err error, streamName string, errorLogged *bool) {
	if !*errorLogged {
		utils.Logger.ErrorF("failed to start %s: %v", streamName, err)
		*errorLogged = true
	} else {
		utils.Logger.LogF(100, "failed to start %s: %v", streamName, err)
	}
}
