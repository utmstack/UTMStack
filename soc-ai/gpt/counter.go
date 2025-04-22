package gpt

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

func NumTokensFromMessage(message string, model string) (numTokens int, err error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return -1, fmt.Errorf("encoding for model: %v", err)
	}

	var tokensPerMessage, tokensPerName int
	switch model {
	case "gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0314",
		"gpt-4-32k-0314",
		"gpt-4-0613",
		"gpt-4-32k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4
		tokensPerName = -1
	default:
		if strings.Contains(model, "gpt-3.5-turbo") {
			return NumTokensFromMessage(message, "gpt-3.5-turbo-0613")
		} else if strings.Contains(model, "gpt-4") {
			return NumTokensFromMessage(message, "gpt-4-0613")
		} else {
			return -1, fmt.Errorf("num_tokens_from_messages() is not implemented for model %s. See https://github.com/openai/openai-python/blob/main/chatml.md for information on how messages are converted to tokens", model)
		}
	}

	numTokens += tokensPerMessage
	numTokens += len(tkm.Encode(message, nil, nil))
	numTokens += tokensPerName

	return numTokens, nil
}
