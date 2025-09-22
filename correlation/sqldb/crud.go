package sqldb

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

func UpdateStatistics(i, s, t string, c int64) {
	ping()

	var err error

	_, err = db.Exec(`INSERT INTO public.utm_asset_metrics
	(id, asset_name, metric, amount) VALUES ($1, $2, $3, $4)
	ON CONFLICT (asset_name, metric)
	DO UPDATE SET amount = public.utm_asset_metrics.amount + $4`, i, s, t, c)

	if err != nil {
		catcher.Error("Error updating statistics for datasource", err, map[string]any{"datasource": s})
	}

	timestamp := time.Now().UTC().Unix()

	ping()

	_, err = db.Exec(`INSERT INTO public.utm_data_input_status
	(id, source, data_type, timestamp, median) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id)
	DO UPDATE SET timestamp=$4, median=$5`, i, s, t, timestamp, int64(10800))

	if err != nil {
		catcher.Error("Error updating status for datasource", err, map[string]any{"datasource": s})
	}
}
