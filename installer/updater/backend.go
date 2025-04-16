package updater

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

func getInstanceInfoFromBackend() InstanceInfo {
	var instanceInfo InstanceInfo

	for {
		time.Sleep(30 * time.Second)
		resp, status, err := utils.DoReq[[]ConfigBackend](
			fmt.Sprintf(config.BackendConfigEndpoint, 6),
			nil,
			http.MethodGet,
			map[string]string{"Utm-Internal-Key": config.GetConfig().InternalKey},
		)
		if err != nil || status != http.StatusOK {
			config.Logger().Info("instance info not ready yet, retrying after error: %v", err)
			continue
		}
		if len(resp) <= 0 {
			config.Logger().Info("instance info not ready yet, retrying after empty response")
			continue
		}

		for _, v := range resp {
			switch v.ConfParamShort {
			case "utmstack.instance.organization":
				instanceInfo.Name = v.ConfParamValue
			case "utmstack.instance.country":
				instanceInfo.Country = v.ConfParamValue
			case "utmstack.instance.contact_email":
				instanceInfo.Email = v.ConfParamValue
			}
		}

		if instanceInfo.Name == "" || instanceInfo.Country == "" || instanceInfo.Email == "" {
			config.Logger().Info("instance info not ready yet, retrying after incomplete data")
			continue
		}

		break
	}

	return instanceInfo
}

func getInstanceAuthFromBackend() (string, error) {
	auth := ""
	resp, status, err := utils.DoReq[[]ConfigBackend](
		fmt.Sprintf(config.BackendConfigEndpoint, 6),
		nil,
		http.MethodGet,
		map[string]string{"Utm-Internal-Key": config.GetConfig().InternalKey},
	)
	if err != nil || status != http.StatusOK {
		return auth, fmt.Errorf("error getting instance auth from backend: status code: %d, error: %v", status, err)
	}

	if len(resp) <= 0 {
		return auth, fmt.Errorf("error getting instance auth from backend: empty response")
	}

	for _, v := range resp {
		if v.ConfParamShort == "utmstack.instance.auth" {
			auth = v.ConfParamValue
			break
		}
	}

	return auth, nil
}

func getLicenseFromBackend() (string, error) {
	resp, status, err := utils.DoReq[[]ConfigBackend](
		fmt.Sprintf(config.BackendConfigEndpoint, 7),
		nil,
		http.MethodGet,
		map[string]string{"Utm-Internal-Key": config.GetConfig().InternalKey},
	)
	if err != nil || status != http.StatusOK {
		return "", fmt.Errorf("error getting license from backend: status code: %d, error: %v", status, err)
	}

	return resp[0].ConfParamValue, nil
}

func updateInstanceInfoInBackend(instanceInfo InstanceInfo) error {
	jsonData, err := json.Marshal(instanceInfo)
	if err != nil {
		return fmt.Errorf("error marshalling instance info: %v", err)
	}

	instanceInfoBase64 := base64.StdEncoding.EncodeToString(jsonData)

	resp, status, err := utils.DoReq[[]ConfigBackend](
		fmt.Sprintf(config.BackendConfigEndpoint, 6),
		nil,
		http.MethodGet,
		map[string]string{"Utm-Internal-Key": config.GetConfig().InternalKey},
	)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("error getting instance info from backend: status code: %d, error: %v", status, err)
	}
	if len(resp) <= 0 {
		return fmt.Errorf("error getting instance info from backend: empty response")
	}

	for i, v := range resp {
		if v.ConfParamShort == "utmstack.instance.data" {
			resp[i].ConfParamValue = instanceInfoBase64
		}
	}

	jsonData, err = json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error marshalling updated instance info: %v", err)
	}

	_, status, err = utils.DoReq[string](
		fmt.Sprintf(config.BackendConfigEndpoint, 6),
		jsonData,
		http.MethodPut,
		map[string]string{"Utm-Internal-Key": config.GetConfig().InternalKey},
	)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("error updating instance info in backend: status code: %d, error: %v", status, err)
	}

	return nil
}
