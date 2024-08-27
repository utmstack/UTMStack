package agent

import (
	"context"
	sync "sync"

	"github.com/utmstack/UTMStack/agent-manager/config"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ModulesServ *ModulesService
	modulesOnce sync.Once
)

type ModulesService struct {
	UnimplementedModuleConfigServiceServer
}

func InitModulesService() *ModulesService {
	modulesOnce.Do(func() {
		ModulesServ = &ModulesService{}
	})
	return ModulesServ
}

func (g *ModulesService) IsModuleEnabled(ctx context.Context, in *ModuleEnabledRequest) (*ModuleEnabled, error) {
	module := in.GetModule()
	if module == "" {
		return nil, status.Error(codes.InvalidArgument, "module is required")
	}

	client := utmconf.NewUTMClient(config.GetInternalKey(), config.GetPanelServiceName())
	moduleConfig, err := client.GetUTMConfig(enum.UTMModule(module))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &ModuleEnabled{Enabled: moduleConfig.ModuleActive}, nil
}
