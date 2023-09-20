package search

import (
	"fmt"
	"time"
)

func IndexBuilder(name, timestamp string) (string, error) {
	st, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return "", err
	}

	fst := st.Format("2006.01.02")

	index := fmt.Sprintf(
		"%s-%s",
		name,
		fst,
	)
	return index, nil
}
