package utils

import "encoding/base64"

func GenerateAuthCode(apiKey string) string {
	loginString := apiKey + ":"
	encodedBytes := base64.StdEncoding.EncodeToString([]byte(loginString))
	authCode := "Basic " + encodedBytes
	return authCode
}
