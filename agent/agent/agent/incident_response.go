package agent

import (
	"context"
	"io"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IncidentResponseStream(client AgentServiceClient, ctx context.Context, cnf *config.Config) {
	path := utils.GetMyPath()

	connectionTime := 0 * time.Second
	reconnectDelay := config.InitialReconnectDelay
	var connErrMsgWritten bool

	for {
		if connectionTime >= config.MaxConnectionTime {
			connectionTime = 0 * time.Second
			reconnectDelay = config.InitialReconnectDelay
			continue
		}

		stream, err := client.AgentStream(ctx)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to start AgentStream: %v", err)
				connErrMsgWritten = true
			}

			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, config.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, config.MaxReconnectDelay)
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
				utils.Logger.ErrorF("error receiving command from server: %v", err)
				break
			}

			switch msg := in.StreamMessage.(type) {
			case *BidirectionalStream_Command:
				// Handle the received command
				err = commandProcessor(path, stream, cnf, []string{msg.Command.Command, in.GetCommand().CmdId})
				if err == io.EOF {
					break
				}
				if err != nil {
					utils.Logger.ErrorF("failed to send result to server: %v", err)
				}
			}
		}
	}
}

func commandProcessor(path string, stream AgentService_AgentStreamClient, cnf *config.Config, commandPair []string) error {
	var result string
	var errB bool

	utils.Logger.Info("Received command: %s", commandPair[0])

	switch runtime.GOOS {
	case "windows":
		result, errB = utils.ExecuteWithResult("cmd.exe", path, "/C", commandPair[0])
	case "linux":
		result, errB = utils.ExecuteWithResult("sh", path, "-c", commandPair[0])
	default:
		utils.Logger.Fatal("unsupported operating system: %s", runtime.GOOS)
	}

	if errB {
		utils.Logger.ErrorF("error executing command %s: %s", commandPair[0], result)
	} else {
		utils.Logger.Info("Result when executing the command %s: %s", commandPair[0], result)
	}

	// Send the result back to the server
	if err := stream.Send(&BidirectionalStream{
		StreamMessage: &BidirectionalStream_Result{
			Result: &CommandResult{Result: result, AgentKey: cnf.AgentKey, ExecutedAt: timestamppb.Now(), CmdId: commandPair[1]},
		},
	}); err != nil {
		return err
	} else {
		utils.Logger.Info("Result sent to server successfully!!!")
	}
	return nil
}
