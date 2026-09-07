package agent

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IncidentResponseStream(cnf *config.Config, ctx context.Context) {
	ensureConfigStateLoaded()

	path := fs.GetExecutablePath()
	var connErrLogged, streamErrLogged bool

	for {
		if ctx.Err() != nil {
			utils.Logger.Info("AgentStream stopping due to context cancellation")
			return
		}

		connection, err := GetAgentManagerConnection(cnf)
		if err != nil {
			LogConnectionError(err, "Agent Manager", &connErrLogged)
			if !sleepOrDone(ctx, timeToSleep) {
				return
			}
			continue
		}

		client := NewAgentServiceClient(connection)
		stream, err := client.AgentStream(ctx)
		if err != nil {
			LogStreamError(err, "AgentStream", &connErrLogged)
			if !sleepOrDone(ctx, timeToSleep) {
				return
			}
			continue
		}

		connErrLogged = false
		serveAgentStream(ctx, stream, path, cnf, &streamErrLogged)
	}
}

func serveAgentStream(ctx context.Context, stream AgentService_AgentStreamClient, path string, cnf *config.Config, streamErrLogged *bool) {
	sender := &lockedStream{stream: stream}

	beatCtx, stopBeating := context.WithCancel(ctx)
	defer stopBeating()
	go sendHeartbeats(beatCtx, sender)
	go sendConfigStateReports(beatCtx, sender)

	for {
		in, err := stream.Recv()
		if err != nil {
			HandleGRPCStreamError(err, "error receiving command from server", streamErrLogged)
			return
		}

		switch msg := in.StreamMessage.(type) {
		case *BidirectionalStream_Command:
			err = commandProcessor(path, sender, cnf, msg.Command.Command, msg.Command.CmdId, msg.Command.Shell)
			if err != nil {
				HandleGRPCStreamError(err, "error sending result to server", streamErrLogged)
				return
			}
		case *BidirectionalStream_ConfigUpdate:
			applyConfigUpdate(msg.ConfigUpdate)
		}
		*streamErrLogged = false
	}
}

func sendHeartbeats(ctx context.Context, sender resultSender) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := sender.Send(&BidirectionalStream{
				StreamMessage: &BidirectionalStream_Heartbeat{Heartbeat: &Heartbeat{}},
			})
			if err != nil {
				utils.Logger.LogF(100, "heartbeat not delivered: %v", err)
				return
			}
		}
	}
}

type resultSender interface {
	Send(*BidirectionalStream) error
}

type lockedStream struct {
	mu     sync.Mutex
	stream AgentService_AgentStreamClient
}

func (l *lockedStream) Send(m *BidirectionalStream) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stream.Send(m)
}

func commandProcessor(path string, stream resultSender, cnf *config.Config, command, cmdId, shell string) error {
	var result string
	var errB bool

	utils.Logger.LogF(100, "Received command: %s (shell: %s)", command, shell)
	if cnf.NoRemoteControl {
		const refused = "refused: this agent was installed with no-remote-control"
		utils.Logger.ErrorF("%s (command: %s)", refused, command)
		return stream.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_Result{
				Result: &CommandResult{
					Result:     refused,
					AgentId:    strconv.Itoa(int(cnf.AgentID)),
					ExecutedAt: timestamppb.Now(),
					CmdId:      cmdId,
				},
			},
		})
	}

	switch runtime.GOOS {
	case "windows":
		if shell == "powershell" {
			result, errB = utils.ExecuteWithResult("powershell.exe", path, "-Command", command)
		} else {
			// Default to cmd.exe (also handles shell == "" or shell == "cmd")
			result, errB = utils.ExecuteWithResult("cmd.exe", path, "/C", command)
		}
	case "linux", "darwin":
		if shell == "bash" {
			result, errB = utils.ExecuteWithResult("bash", path, "-c", command)
		} else {
			// Default to sh (also handles shell == "" or shell == "sh")
			result, errB = utils.ExecuteWithResult("sh", path, "-c", command)
		}
	default:
		utils.Logger.ErrorF("unsupported operating system: %s", runtime.GOOS)
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if errB {
		utils.Logger.ErrorF("error executing command %s: %s", command, result)
	} else {
		utils.Logger.LogF(100, "Result when executing the command %s: %s", command, result)
	}

	if err := stream.Send(&BidirectionalStream{
		StreamMessage: &BidirectionalStream_Result{
			Result: &CommandResult{Result: result, AgentId: strconv.Itoa(int(cnf.AgentID)), ExecutedAt: timestamppb.Now(), CmdId: cmdId},
		},
	}); err != nil {
		return err
	}

	utils.Logger.LogF(100, "Result sent to server successfully")
	return nil
}
