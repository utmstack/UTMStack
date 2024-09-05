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

	go_sdk "github.com/threatwinds/go-sdk"
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
	intKey := GetConfig().InternalKey
	panelServ := GetConfig().Backend
	client := UTMStackConfigurationClient.NewUTMClient(intKey, "http://"+panelServ)
	for {
		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.BITDEFENDER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				go_sdk.Logger().ErrorF("error getting configuration of the Bitdefender module: %v", err)
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
					go_sdk.Logger().Info("new configuration found: groupName: %s, master: %s, CompanyIDs: %s", group.GroupName, group.Configurations[2].ConfValue, group.Configurations[3].ConfValue)
					if err := confBDGZApiPush(group, "sendConf"); err != nil {
						go_sdk.Logger().ErrorF("error sending configuration")
						continue
					}
					time.Sleep(15 * time.Second)
					if err := confBDGZApiPush(group, "getConf"); err != nil {
						go_sdk.Logger().ErrorF("error getting configuration")
						continue
					}
					if err := confBDGZApiPush(group, "sendTest"); err != nil {
						go_sdk.Logger().ErrorF("error sending test event")
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
			go_sdk.Logger().ErrorF("%v", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		defer response.Body.Close()
		go_sdk.Logger().Info("Status: %s", response.Status)
		myBody, _ := io.ReadAll(response.Body)
		go_sdk.Logger().Info(string(myBody))

		if operation == "sendConf" {
			// Check if config was sent correctly
			regex := regexp.MustCompile(`result":true`)
			match := regex.Match([]byte(string(myBody)))
			if match {
				go_sdk.Logger().Info("Configuration sent correctly")
			}
		}
		return nil
	}
	return fmt.Errorf("error sending configuration")
}

func sendPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	go_sdk.Logger().Info("Sending configuration...")
	byteTemplate := getTemplateSetPush(config)
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		go_sdk.Logger().ErrorF("error when marshaling the request body to send the configuration: %v", err)
		return nil, err
	}
	return sendRequest(body, config)
}

func getPushEventSettings(config types.ModuleGroup) (*http.Response, error) {
	go_sdk.Logger().Info("Checking configuration...")
	byteTemplate := getTemplateGet()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		go_sdk.Logger().ErrorF("error when marshaling the request body to send the configuration: %v", err)
		return nil, err
	}
	return sendRequest(body, config)
}

func sendTestPushEvent(config types.ModuleGroup) (*http.Response, error) {
	go_sdk.Logger().Info("Sending Event Test...")
	byteTemplate := getTemplateTest()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		go_sdk.Logger().ErrorF("error when marshaling the request body to send the configuration: %v", err)
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
