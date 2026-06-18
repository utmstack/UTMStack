package utils

import (
	"encoding/base64"
)

func GenerateKey(baseKey string) ([]byte, error) {
	info, err := GetOsInfo()
	if err != nil {
		return nil, Logger.ErrorF("error getting os info: %v", err)
	}

	data := []byte(info.Hostname + info.Mac + info.OsType)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(baseKey + base64Key), nil
}

func GenerateKeyByUUID(baseKey string, uuid string) ([]byte, error) {
	data := []byte(baseKey + uuid)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(base64Key), nil
}
