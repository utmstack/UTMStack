package utils

import "encoding/base64"

func GenerateAuthCode(apiKey string) string {
	loginString := apiKey + ":"
	encodedBytes := base64.StdEncoding.EncodeToString([]byte(loginString))
	encodedUserPassSequence := string(encodedBytes[:])
	authCode := "Basic " + encodedUserPassSequence
	return authCode
}
