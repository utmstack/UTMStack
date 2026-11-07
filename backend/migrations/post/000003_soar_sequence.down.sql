DROP INDEX IF EXISTS idx_soar_execution_run_node;
ALTER TABLE soar_executions DROP COLUMN IF EXISTS sequence;
CREATE UNIQUE INDEX IF NOT EXISTS idx_soar_execution_run_node
    ON soar_executions (flow_run_id, node_id, depth);
