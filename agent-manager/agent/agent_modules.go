package agent

import (
	"errors"
	"io"

	"github.com/utmstack/UTMStack/agent-manager/enum"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Grpc) GetAgentConfig() {
}

func (s *Grpc) AgentModuleUpdateStream(stream AgentConfigService_AgentModuleUpdateStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return errors.New("unable to get the context")
		}

		module, err := agentModuleService.FindByID(uint(req.AgentModuleId))
		if err != nil {
			return errors.New("unable to get the module related")
		}
		// Get the authorization token from metadata
		agentKey := md.Get("Agent-key")[0]
		agentStream, ok := s.AgentStreamMap[agentKey]
		if !ok {
			return status.Errorf(codes.NotFound, "agent not found")
		}
		var value = req.ConfValue
		if req.ConfDatatype == enum.PasswordType {
			value, err = util.DecryptValue(value)
			if err != nil {
				return status.Errorf(codes.Internal, "unable to decrypt config")
			}
		}

		err = agentStream.stream.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_Config{
				Config: &UpdateAgentModule{
					AgentModuleShort: module.ShortName,
					ConfKey:          req.ConfKey,
					ConfValue:        value,
				},
			},
		})
		if err != nil {
			return err
		}

	}
	return nil
}
