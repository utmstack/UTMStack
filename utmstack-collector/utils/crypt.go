package utils

import (
	"encoding/base64"
)

func GenerateKey(baseKey string) ([]byte, error) {
	if baseKey == "" {
		return nil, Logger.ErrorF("build secret is not set: this binary was built without the REPLACE_KEY ldflag, so the at-rest key would carry no secret")
	}

	info, err := GetOsInfo()
	if err != nil {
		return nil, Logger.ErrorF("error getting os info: %v", err)
	}

	data := []byte(info.Hostname + info.Mac + info.OsType)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(baseKey + base64Key), nil
}

func GenerateKeyByUUID(baseKey string, uuid string) ([]byte, error) {
	if baseKey == "" {
		return nil, Logger.ErrorF("build secret is not set: this binary was built without the REPLACE_KEY ldflag, so the at-rest key would carry no secret")
	}

	data := []byte(baseKey + uuid)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(base64Key), nil
}
