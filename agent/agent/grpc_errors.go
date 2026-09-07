package agent

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGRPCStreamError(err error, msg string, errorLogged *bool) {
	if strings.Contains(err.Error(), "EOF") {
		utils.Logger.LogF(100, "%s: %v", msg, err)
		time.Sleep(timeToSleep)
		return
	}

	if st, ok := status.FromError(err); ok &&
		(st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
		utils.Logger.LogF(100, "%s: %v", msg, err)
		time.Sleep(timeToSleep)
		return
	}

	logError(err, msg, errorLogged)
	time.Sleep(timeToSleep)
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

func sleepOrDone(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
