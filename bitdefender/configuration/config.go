package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/bitdefender/constants"
	"github.com/utmstack/UTMStack/bitdefender/utils"
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
	intKey := constants.GetInternalKey()
	panelServ := constants.GetPanelServiceName()
	client := UTMStackConfigurationClient.NewUTMClient(intKey, "http://"+panelServ)
	for {
		time.Sleep(delayCheckConfig)

		if err := utils.ConnectionChecker(constants.URL_CHECK_CONNECTION); err != nil {
			catcher.Error("Failed to establish connection", err, nil)
		}

		tempModuleConfig, err := client.GetUTMConfig(enum.BITDEFENDER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				catcher.Error("error getting configuration of the Bitdefender module", err, nil)
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
					catcher.Info("new configuration found", map[string]any{"groupName": group.GroupName, "master": group.Configurations[2].ConfValue, "CompanyIDs": group.Configurations[3].ConfValue})
					if err := confBDGZApiPush(group, "sendConf"); err != nil {
						catcher.Error("error sending configuration", err, nil)
						continue
					}
					time.Sleep(15 * time.Second)
					if err := confBDGZApiPush(group, "getConf"); err != nil {
						catcher.Error("error getting configuration", err, nil)
						continue
					}
					if err := confBDGZApiPush(group, "sendTest"); err != nil {
						catcher.Error("error sending test event", err, nil)
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

func confBDGZApiPush(config types.ModuleGroup, operation string) error {
	operationFunc := map[string]func(types.ModuleGroup) (*http.Response, error){
		"sendConf": sendPushEventSettings,
		"getConf":  getPushEventSettings,
		"sendTest": sendTestPushEvent,
	}

	fn, ok := operationFunc[operation]
	if !ok {
		return fmt.Errorf("wrong operation")
	}

	for i := 0; i < 5; i++ {
		response, err := fn(config)
		if err != nil {
			catcher.Error("error sending configuration", err, nil)
			time.Sleep(1 * time.Minute)
			continue
		}
		defer response.Body.Close()
		catcher.Info("Status", map[string]any{"status": response.Status})
		myBody, _ := io.ReadAll(response.Body)
		catcher.Info("Response Body", map[string]any{"body": string(myBody)})

		if operation == "sendConf" {
			regex := regexp.MustCompile(`result":true`)
			match := regex.Match([]byte(string(myBody)))
			if match {
				catcher.Info("Configuration sent correctly", nil)
			}
		}
		return nil
	}
	return fmt.Errorf("error sending configuration")
}

func sendPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	catcher.Info("Sending configuration...", nil)
	byteTemplate := getTemplateSetPush(config)
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		catcher.Error("error when marshaling the request body to send the configuration", err, nil)
		return nil, err
	}
	return sendRequest(body, config)
}

func getPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	catcher.Info("Checking configuration...", nil)
	byteTemplate := getTemplateGet()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		catcher.Error("error when marshaling the request body to send the configuration", err, nil)
		return nil, err
	}
	return sendRequest(body, config)
}

func sendTestPushEvent(config types.ModuleGroup) (*http.Response, error) {
	catcher.Info("Sending Event Test...", nil)
	byteTemplate := getTemplateTest()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		catcher.Error("error when marshaling the request body to send the configuration", err, nil)
		return nil, err
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
