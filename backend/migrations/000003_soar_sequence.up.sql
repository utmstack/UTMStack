-- Per-trigger execution: a node may fire multiple times at the same depth,
-- so the unique key gains a sequence discriminator.
DROP INDEX IF EXISTS idx_soar_execution_run_node;
ALTER TABLE soar_executions ADD COLUMN IF NOT EXISTS sequence integer NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_soar_execution_run_node
    ON soar_executions (flow_run_id, node_id, depth, sequence);
