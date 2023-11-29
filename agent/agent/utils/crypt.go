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
