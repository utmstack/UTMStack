package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/bdgz/constants"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"

	UTMStackConfigurationClient "github.com/utmstack/config-client-go"
)

const delayCheckConfig = 1 * time.Minute

var configsSent = []string{}

// ConfigureModules updates the module configuration every minute.
func ConfigureModules(cnf *types.ConfigurationSection, mutex *sync.Mutex, h *holmes.Logger) {
	intKey := constants.GetInternalKey()
	panelServ := constants.GetPanelServiceName()
	client := UTMStackConfigurationClient.NewUTMClient(intKey, "http://"+panelServ)
	for {
		time.Sleep(delayCheckConfig)
		// Get Bitdefender module configs
		tempModuleConfig, err := client.GetUTMConfig(enum.BITDEFENDER)
		if err != nil {
			if (err.Error() != "") && (err.Error() != " ") {
				h.Error("error getting configuration of the Bitdefender module: %v", err)
			}
			continue
		}
		mutex.Lock()
		*cnf = *tempModuleConfig
		mutex.Unlock()

		// Configure the group to send data to the syslog server if it is not already configured
		for _, group := range (*cnf).ConfigurationGroups {
			var configured bool
			for _, name := range configsSent {
				if name == group.GroupName {
					configured = true
				}
			}
			if !configured {
				h.Info("new configuration found: groupName: %s, master: %s, CompanyIDs: %s", group.GroupName, group.Configurations[2].ConfValue, group.Configurations[3].ConfValue)
				if err := confBDGZApiPush(group, "sendConf", h); err != nil {
					h.Error("error sending configuration")
					continue
				}
				time.Sleep(15 * time.Second)
				if err := confBDGZApiPush(group, "getConf", h); err != nil {
					h.Error("error getting configuration")
					continue
				}
				if err := confBDGZApiPush(group, "sendTest", h); err != nil {
					h.Error("error sending test event")
					continue
				}
				configsSent = append(configsSent, group.GroupName)
			}
		}
	}
}

// confBDGZApiPush configures the Bitdefender API.
// It checks to check that everything has been configured correctly.
// Send test logs
func confBDGZApiPush(config types.ModuleGroup, operation string, h *holmes.Logger) error {
	operationFunc := map[string]func(types.ModuleGroup, *holmes.Logger) (*http.Response, error){
		"sendConf": sendPushEventSettings,
		"getConf":  getPushEventSettings,
		"sendTest": sendTestPushEvent,
	}

	fn, ok := operationFunc[operation]
	if !ok {
		return fmt.Errorf("wrong operation")
	}

	for i := 0; i < 5; i++ {
		response, err := fn(config, h)
		if err != nil {
			h.Error("%v", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		defer response.Body.Close()
		h.Info("Status: ", response.Status)
		myBody, _ := io.ReadAll(response.Body)
		h.Info(string(myBody))

		if operation == "sendConf" {
			// Check if config was sent correctly
			regex := regexp.MustCompile(`result":true`)
			match := regex.Match([]byte(string(myBody)))
			if match {
				h.Info("Configuration sent correctly")
			}
		}
		return nil
	}
	return fmt.Errorf("error sending configuration")
}

// setPushEventSettings sends the configuration to the Bitdefender API
func sendPushEventSettings(config types.ModuleGroup, h *holmes.Logger) (*http.Response, error) {
	h.Info("Sending configuration...")
	byteTemplate := getTemplateSetPush(config)
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		h.Error("error when marshaling the request body to send the configuration: ", err)
		return nil, err
	}
	return sendRequest(body, config)
}

// getPushEventSettings gets the Bitdefender API settings
func getPushEventSettings(config types.ModuleGroup, h *holmes.Logger) (*http.Response, error) {
	h.Info("Checking configuration...")
	byteTemplate := getTemplateGet()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		h.Error("error when marshaling the request body to send the configuration: ", err)
		return nil, err
	}
	return sendRequest(body, config)
}

// sendTestPushEvent sends a test event to the connector
func sendTestPushEvent(config types.ModuleGroup, h *holmes.Logger) (*http.Response, error) {
	h.Info("Sending Event Test...")
	byteTemplate := getTemplateTest()
	body, err := json.Marshal(byteTemplate)
	if err != nil {
		h.Error("error when marshaling the request body to send the configuration: ", err)
		return nil, err
	}
	return sendRequest(body, config)
}
