package utils

import (
	"encoding/base64"
	"fmt"
)

func GenerateKey(REPALCE_KEY string) ([]byte, error) {
	info, err := GetOsInfo()
	if err != nil {
		return nil, fmt.Errorf("error getting os info: %v", err)
	}

	data := []byte(info.Hostname + info.Mac + info.OsType)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(REPALCE_KEY + base64Key), nil
}

func GenerateKeyByUUID(REPLACE_KEY string, uuid string) ([]byte, error) {
	data := []byte(REPLACE_KEY + uuid)
	base64Key := base64.StdEncoding.EncodeToString(data)
	return []byte(base64Key), nil
}
