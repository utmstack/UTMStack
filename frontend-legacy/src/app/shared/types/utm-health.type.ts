export class UtmHealthType {
  status: 'UP' | 'DOWN';
  components: {
    db: {
      status: 'UP' | 'DOWN',
      details: {
        database: string,
        validationQuery: boolean
      }
    },
    diskSpace: {
      status: 'UP' | 'DOWN',
      details: {
        total: number,
        free: number,
        threshold: number,
        exists: true
      }
    },
    elasticsearch: {
      status: 'UP' | 'DOWN';
      details: {
        cluster_name: string,
        status: string,
        timed_out: false,
        number_of_nodes: number,
        number_of_data_nodes: number,
        active_primary_shards: number,
        active_shards: number,
        relocating_shards: number,
        initializing_shards: number,
        unassigned_shards: number,
        delayed_unassigned_shards: number,
        number_of_pending_tasks: number,
        number_of_in_flight_fetch: number,
        task_max_waiting_in_queue_millis: number,
        active_shards_percent_as_number: number
      };
    };
    livenessState: {
      status: 'UP' | 'DOWN'
    };
    ping: {
      status: 'UP' | 'DOWN'
    };
    readinessState: {
      status: 'UP' | 'DOWN'
    }
  };

}
