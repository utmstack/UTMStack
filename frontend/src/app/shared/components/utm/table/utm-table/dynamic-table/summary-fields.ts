export const SUMMARY_COLUMNS = [
  {
    pattern: 'v11-log-wineventlog-*',
    fields: [
      {key: {field: 'log.eventName'}, title: 'Event'},
      {key: {field: 'dataType'}, title: 'Data Type'},
      {key: {field: 'dataSource'}, title: 'Data Source'},
    ]
  },
  {
    pattern: 'v11-alert-*',
    fields: [
      {key: {field: 'description'}, title: 'Description'},
      {key: {field: 'dataType'}, title: 'Data Type'},
      {key: {field: 'dataSource'}, title: 'Data Source'},
    ]
  },
  {
    pattern: 'v11-log-*',
    fields: [
      {key: {field: 'tenantName'}, title: 'Tenant'},
      {key: {field: 'dataType'}, title: 'Data Type'},
      {key: {field: 'dataSource'}, title: 'Data Source'},
    ]
  }
];

