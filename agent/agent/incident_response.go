package agent

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/conn"
	"github.com/utmstack/UTMStack/agent/utils"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IncidentResponseStream(cnf *config.Config, ctx context.Context) {
	path := utils.GetMyPath()
	var connErrMsgWritten, errorLogged bool

	for {
		connection, err := conn.GetAgentManagerConnection(cnf)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("error connecting to Agent Manager: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(timeToSleep)
			continue
		}

		client := NewAgentServiceClient(connection)
		stream, err := client.AgentStream(ctx)
		if err != nil {
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to start AgentStream: %v", err)
				connErrMsgWritten = true
			} else {
				utils.Logger.LogF(100, "failed to start AgentStream: %v", err)
			}
			time.Sleep(timeToSleep)
			continue
		}

		connErrMsgWritten = false

		for {
			in, err := stream.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					utils.Logger.LogF(100, "error receiving command from server: %v", err)
					time.Sleep(timeToSleep)
					break
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					if !errorLogged {
						utils.Logger.ErrorF("error receiving command from server: %v", err)
						errorLogged = true
					} else {
						utils.Logger.LogF(100, "error receiving command from server: %v", err)
					}
					time.Sleep(timeToSleep)
					break
				} else {
					if !errorLogged {
						utils.Logger.ErrorF("error receiving command from server: %v", err)
						errorLogged = true
					} else {
						utils.Logger.LogF(100, "error receiving command from server: %v", err)
					}
					time.Sleep(timeToSleep)
					continue
				}
			}

			switch msg := in.StreamMessage.(type) {
			case *BidirectionalStream_Command:
				err = commandProcessor(path, stream, cnf, []string{msg.Command.Command, in.GetCommand().CmdId})
				if err != nil {
					if strings.Contains(err.Error(), "EOF") {
						utils.Logger.LogF(100, "error sending result to server: %v", err)
						time.Sleep(timeToSleep)
						break
					}
					st, ok := status.FromError(err)
					if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
						if !errorLogged {
							utils.Logger.ErrorF("error sending result to server: %v", err)
							errorLogged = true
						} else {
							utils.Logger.LogF(100, "error sending result to server: %v", err)
						}
						time.Sleep(timeToSleep)
						break
					} else {
						if !errorLogged {
							utils.Logger.ErrorF("error sending result to server: %v", err)
							errorLogged = true
						} else {
							utils.Logger.LogF(100, "error sending result to server: %v", err)
						}
						time.Sleep(timeToSleep)
						continue
					}
				}
			}
			errorLogged = false
		}
	}
}

func commandProcessor(path string, stream AgentService_AgentStreamClient, cnf *config.Config, commandPair []string) error {
	var result string
	var errB bool

	utils.Logger.LogF(100, "Received command: %s", commandPair[0])

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
		utils.Logger.LogF(100, "Result when executing the command %s: %s", commandPair[0], result)
	}

	if err := stream.Send(&BidirectionalStream{
		StreamMessage: &BidirectionalStream_Result{
			Result: &CommandResult{Result: result, AgentKey: cnf.AgentKey, ExecutedAt: timestamppb.Now(), CmdId: commandPair[1]},
		},
	}); err != nil {
		return err
	} else {
		utils.Logger.LogF(100, "Result sent to server successfully!!!")
	}
	return nil
}
