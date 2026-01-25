package workers

import (
	"context"
	"runtime"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/client"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/evaluator"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/scheduler"
)

func StartWorkers(ctx context.Context, backend *client.BackendClient) {
	numWorkers := 2 * runtime.NumCPU()

	for i := 0; i < numWorkers; i++ {
		go func(id int) {

			eval := evaluator.NewEvaluator(backend)

			for cfg := range scheduler.Jobs {

				// Log opcional
				catcher.Info("Worker evaluating control", map[string]any{
					"worker":  id,
					"control": cfg.ID,
				})

				// Ejecutar evaluación real
				_, err := eval.Evaluate(ctx, cfg)
				if err != nil {
					// catcher.New("evaluation failed").
					//     Set("worker", id).
					//     Set("control", cfg.ID).
					//     SetError(err).
					//     Log()
					continue
				}

				// Aquí luego enviarás el resultado al sender
				// sender.Send(result)
			}
		}(i)
	}
}
