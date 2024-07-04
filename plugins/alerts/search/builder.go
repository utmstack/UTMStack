package search

import (
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/logger"
)

func IndexBuilder(name, timestamp string) (string, *logger.Error) {
	st, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return "", helpers.Logger().ErrorF(err.Error())
	}

	fst := st.Format("2006.01.02")

	index := fmt.Sprintf(
		"%s-%s",
		name,
		fst,
	)
	return index, nil
}
