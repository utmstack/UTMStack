package configuration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	gosdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"

	UTMStackConfigurationClient "github.com/utmstack/config-client-go"
)

type ModuleConfig struct {
	ConnectionKey string
	AccessUrl     string
	MasterIp      string
	CompaniesIDs  []string
}

const delayCheckConfig = 30 * time.Second

var configsSent = make(map[string]ModuleConfig)

func ConfigureModules(cnf *types.ConfigurationSection, mutex *sync.Mutex) {
	for {
		utmConfig := gosdk.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		client := UTMStackConfigurationClient.NewUTMClient(internalKey, backendUrl)

		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.BITDEFENDER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = gosdk.Error("error getting configuration of the Bitdefender module", err, map[string]any{})
			}
			continue
		}
		mutex.Lock()
		*cnf = *tempModuleConfig
		mutex.Unlock()

		for _, group := range (*cnf).ConfigurationGroups {
			isNecessaryConfig := compareConfigs(configsSent, group)
			if isNecessaryConfig {
				if !araAnyEmpty(group.Configurations[0].ConfValue, group.Configurations[1].ConfValue, group.Configurations[2].ConfValue, group.Configurations[3].ConfValue) {
					if err := apiPush(group, "sendConf"); err != nil {
						_ = gosdk.Error("error sending configuration", err, map[string]any{})
						continue
					}
					time.Sleep(15 * time.Second)
					if err := apiPush(group, "getConf"); err != nil {
						_ = gosdk.Error("error getting configuration", err, map[string]any{})
						continue
					}
					if err := apiPush(group, "sendTest"); err != nil {
						_ = gosdk.Error("error sending test event", err, map[string]any{})
						continue
					}

					configsSent[group.GroupName] = ModuleConfig{
						ConnectionKey: group.Configurations[0].ConfValue,
						AccessUrl:     group.Configurations[1].ConfValue,
						MasterIp:      group.Configurations[2].ConfValue,
						CompaniesIDs:  strings.Split(group.Configurations[3].ConfValue, ","),
					}
				}
			}
		}
	}
}

func apiPush(config types.ModuleGroup, operation string) error {
	operationFunc := map[string]func(types.ModuleGroup) (*http.Response, error){
		"sendConf": sendPushEventSettings,
		"getConf":  getPushEventSettings,
		"sendTest": sendTestPushEvent,
	}

	fn, ok := operationFunc[operation]
	if !ok {
		return gosdk.Error("wrong operation", nil, map[string]any{})
	}

	for i := 0; i < 5; i++ {
		response, err := fn(config)
		if err != nil {
			_ = gosdk.Error(fmt.Sprintf("%v", err), err, map[string]any{})
			time.Sleep(1 * time.Minute)
			continue
		}

		func() { _ = response.Body.Close() }()

		return nil
	}

	return gosdk.Error("error sending configuration after 5 retries", nil, map[string]any{})
}

func sendPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	byteTemplate := getTemplateSetPush(config)
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		return nil, gosdk.Error("error when marshaling the request body to send the configuration", err, map[string]any{})
	}
	return sendRequest(body, config)
}

func getPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	byteTemplate := getTemplateGet()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		return nil, gosdk.Error("error when marshaling the request body to get the configuration", err, map[string]any{})
	}
	return sendRequest(body, config)
}

func sendTestPushEvent(config types.ModuleGroup) (*http.Response, error) {
	byteTemplate := getTemplateTest()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		return nil, gosdk.Error("error when marshaling the request body to send the test event", err, map[string]any{})
	}
	return sendRequest(body, config)
}

func araAnyEmpty(strings ...string) bool {
	for _, s := range strings {
		if s == "" || s == " " {
			return true
		}
	}
	return false
}
