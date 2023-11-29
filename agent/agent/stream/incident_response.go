package stream

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/quantfall/holmes"
	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IncidentResponseStream(client pb.AgentServiceClient, ctx context.Context, cnf *configuration.Config, h *holmes.Logger) {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("Failed to get current path: %v", err)
		h.FatalError("Failed to get current path: %v", err)
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
				h.Error("failed to start AgentStream: %v", err)
				connErrMsgWritten = true
			}

			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
			continue
		}

		// Send the authentication response
		authResponse := &pb.BidirectionalStream{
			StreamMessage: &pb.BidirectionalStream_AuthResponse{
				AuthResponse: &pb.AuthResponse{AgentKey: cnf.AgentKey, AgentId: uint64(cnf.AgentID)},
			},
		}

		if err := stream.Send(authResponse); err != nil {
			if !connErrMsgWritten {
				h.Error("failed to send AuthResponse: %v", err)
				connErrMsgWritten = true
			}
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
			time.Sleep(reconnectDelay)
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
				h.Error("error receiving command from server: %v", err)
				break
			}

			switch msg := in.StreamMessage.(type) {
			case *pb.BidirectionalStream_Command:
				// Handle the received command
				err = commandProcessor(h, path, stream, cnf, []string{msg.Command.Command, in.GetCommand().CmdId})
				if err == io.EOF {
					break
				}
				if err != nil {
					h.Error("failed to send result to server: %v", err)
				}
			}
		}
	}
}

func commandProcessor(h *holmes.Logger, path string, stream pb.AgentService_AgentStreamClient, cnf *configuration.Config, commandPair []string) error {
	var result string
	var errB bool

	h.Info("Received command: %s", commandPair[0])

	switch runtime.GOOS {
	case "windows":
		result, errB = utils.ExecuteWithResult("cmd.exe", path, "/C", commandPair[0])
	case "linux":
		result, errB = utils.ExecuteWithResult("sh", path, "-c", commandPair[0])
	default:
		h.FatalError("unsupported operating system: %s", runtime.GOOS)
	}

	if errB {
		h.Error("error executing command %s: %s", commandPair[0], result)
	} else {
		h.Info("Result when executing the command %s: %s", commandPair[0], result)
	}

	// Send the result back to the server
	if err := stream.Send(&pb.BidirectionalStream{
		StreamMessage: &pb.BidirectionalStream_Result{
			Result: &pb.CommandResult{Result: result, AgentKey: cnf.AgentKey, ExecutedAt: timestamppb.Now(), CmdId: commandPair[1]},
		},
	}); err != nil {
		return err
	} else {
		h.Info("Result sent to server successfully!!!")
	}
	return nil
}
