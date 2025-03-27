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
			utils.Logger.ErrorF("Failed to establish connection: %v", err)
		}

		tempModuleConfig, err := client.GetUTMConfig(enum.BITDEFENDER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				utils.Logger.ErrorF("error getting configuration of the Bitdefender module: %v", err)
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
					utils.Logger.Info("new configuration found: groupName: %s, master: %s, CompanyIDs: %s", group.GroupName, group.Configurations[2].ConfValue, group.Configurations[3].ConfValue)
					if err := confBDGZApiPush(group, "sendConf"); err != nil {
						utils.Logger.ErrorF("error sending configuration")
						continue
					}
					time.Sleep(15 * time.Second)
					if err := confBDGZApiPush(group, "getConf"); err != nil {
						utils.Logger.ErrorF("error getting configuration")
						continue
					}
					if err := confBDGZApiPush(group, "sendTest"); err != nil {
						utils.Logger.ErrorF("error sending test event")
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
			utils.Logger.ErrorF("%v", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		defer response.Body.Close()
		utils.Logger.Info("Status: %s", response.Status)
		myBody, _ := io.ReadAll(response.Body)
		utils.Logger.Info("%s", string(myBody))

		if operation == "sendConf" {
			regex := regexp.MustCompile(`result":true`)
			match := regex.Match([]byte(string(myBody)))
			if match {
				utils.Logger.Info("Configuration sent correctly")
			}
		}
		return nil
	}
	return fmt.Errorf("error sending configuration")
}

func sendPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	utils.Logger.Info("Sending configuration...")
	byteTemplate := getTemplateSetPush(config)
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		utils.Logger.ErrorF("error when marshaling the request body to send the configuration: %v", err)
		return nil, err
	}
	return sendRequest(body, config)
}

func getPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	utils.Logger.Info("Checking configuration...")
	byteTemplate := getTemplateGet()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		utils.Logger.ErrorF("error when marshaling the request body to send the configuration: %v", err)
		return nil, err
	}
	return sendRequest(body, config)
}

func sendTestPushEvent(config types.ModuleGroup) (*http.Response, error) {
	utils.Logger.Info("Sending Event Test...")
	byteTemplate := getTemplateTest()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		utils.Logger.ErrorF("error when marshaling the request body to send the configuration: %v", err)
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
