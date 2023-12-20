export const HOT_ROLLOVER_PATH = 'policy.states[index].actions[0].rollover.min_index_age';
export const WARM_ROLLOVER_PATH = 'policy.states[index].transitions[0].conditions.min_index_age';
export const COLD_ROLLOVER_PATH = 'policy.states[index].transitions[0].conditions.min_index_age';

const policyExample = {
  policy: {
    description: 'HotWarmColdDeleteWorkflow', default_state: 'hot', schema_version: 1, states: [
      {
        name: 'hot',
        actions: [{index_priority: {priority: 50}, rollover: {min_index_age: '90d'}}],
        transitions: [{state_name: 'warm'}]
      },
      {
        name: 'warm',
        actions: [{
          index_priority: {priority: 25},
          replica_count: {number_of_replicas: 0},
          force_merge: {max_num_segments: 1}
        }],
        transitions: [{state_name: 'cold', conditions: {min_index_age: '180d'}}]
      },
      {
        name: 'cold',
        actions: [{index_priority: {priority: 0}, read_only: {}}],
        transitions: [{state_name: 'delete', conditions: {min_index_age: '360d'}}]
      }, {name: 'delete', actions: [{delete: {}}]}]
  }
};
