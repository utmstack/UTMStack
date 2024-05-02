package agent

import (
	"context"
	"io"
	"runtime"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IncidentResponseStream(client AgentServiceClient, ctx context.Context, cnf *configuration.Config, h *logger.Logger) {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		h.Fatal("Failed to get current path: %v", err)
	}

	connectionTime := 0 * time.Second
	reconnectDelay := configuration.InitialReconnectDelay
	var connErrMsgWritten bool

	for {
		if connectionTime >= configuration.MaxConnectionTime {
			connectionTime = 0 * time.Second
			reconnectDelay = configuration.InitialReconnectDelay
			continue
		}

		stream, err := client.AgentStream(ctx)
		if err != nil {
			if !connErrMsgWritten {
				h.ErrorF("failed to start AgentStream: %v", err)
				connErrMsgWritten = true
			}

			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
			continue
		}

		connErrMsgWritten = false

		// Handle the bidirectional stream
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// Server closed the stream
				break
			}
			if err != nil {
				h.ErrorF("error receiving command from server: %v", err)
				break
			}

			switch msg := in.StreamMessage.(type) {
			case *BidirectionalStream_Command:
				// Handle the received command
				err = commandProcessor(h, path, stream, cnf, []string{msg.Command.Command, in.GetCommand().CmdId})
				if err == io.EOF {
					break
				}
				if err != nil {
					h.ErrorF("failed to send result to server: %v", err)
				}
			}
		}
	}
}

func commandProcessor(h *logger.Logger, path string, stream AgentService_AgentStreamClient, cnf *configuration.Config, commandPair []string) error {
	var result string
	var errB bool

	h.Info("Received command: %s", commandPair[0])

	switch runtime.GOOS {
	case "windows":
		result, errB = utils.ExecuteWithResult("cmd.exe", path, "/C", commandPair[0])
	case "linux":
		result, errB = utils.ExecuteWithResult("sh", path, "-c", commandPair[0])
	default:
		h.Fatal("unsupported operating system: %s", runtime.GOOS)
	}

	if errB {
		h.ErrorF("error executing command %s: %s", commandPair[0], result)
	} else {
		h.Info("Result when executing the command %s: %s", commandPair[0], result)
	}

	// Send the result back to the server
	if err := stream.Send(&BidirectionalStream{
		StreamMessage: &BidirectionalStream_Result{
			Result: &CommandResult{Result: result, AgentKey: cnf.AgentKey, ExecutedAt: timestamppb.Now(), CmdId: commandPair[1]},
		},
	}); err != nil {
		return err
	} else {
		h.Info("Result sent to server successfully!!!")
	}
	return nil
}
