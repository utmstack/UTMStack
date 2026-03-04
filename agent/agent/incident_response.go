package agent

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IncidentResponseStream(cnf *config.Config, ctx context.Context) {
	path := fs.GetExecutablePath()
	var connErrLogged, streamErrLogged bool

	for {
		connection, err := GetAgentManagerConnection(cnf)
		if err != nil {
			LogConnectionError(err, "Agent Manager", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		client := NewAgentServiceClient(connection)
		stream, err := client.AgentStream(ctx)
		if err != nil {
			LogStreamError(err, "AgentStream", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		connErrLogged = false

	recvLoop:
		for {
			in, err := stream.Recv()
			if err != nil {
				action := HandleGRPCStreamError(err, "error receiving command from server", &streamErrLogged)
				if action == ActionReconnect {
					break recvLoop
				}
				continue
			}

			switch msg := in.StreamMessage.(type) {
			case *BidirectionalStream_Command:
				err = commandProcessor(path, stream, cnf, msg.Command.Command, msg.Command.CmdId, msg.Command.Shell)
				if err != nil {
					action := HandleGRPCStreamError(err, "error sending result to server", &streamErrLogged)
					if action == ActionReconnect {
						break recvLoop
					}
					continue
				}
			}
			streamErrLogged = false
		}
	}
}

func commandProcessor(path string, stream AgentService_AgentStreamClient, cnf *config.Config, command, cmdId, shell string) error {
	var result string
	var errB bool

	utils.Logger.LogF(100, "Received command: %s (shell: %s)", command, shell)

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
