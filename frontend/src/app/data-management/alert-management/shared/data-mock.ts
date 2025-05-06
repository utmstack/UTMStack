import {UtmAlertType} from '../../../shared/types/alert/utm-alert.type';

export const alerts: any[] = [
  {
    parentId: null,
    children: [
      '97a9ab4b-023d-4148-b830-4170dc221caf',
      'bb820193-3a95-4c64-b9d7-ad3a228fe0c7'
    ],
    notes : '',
    description : 'This alert will report each IP from India doing requests to our CDN',
    technique : 'QA',
    reference : [ 'https://google.com' ],
    incidentDetail : {
      createdBy : '',
      observation : '',
      source : '',
      creationDate : ''
    },
    solution : '',
    id : '5d1ef1c6-d725-4843-98f6-06a6b91c394e',
    events : [ {
      severity : 'low',
      log : {
        jsonPayloadQueryType : 'A',
        resourceLabelsLocation : 'global',
        jsonPayloadDns64Translated : false,
        resourceLabelsProjectId : 'utmstack-377415',
        receiveTimestamp : '2025-02-13T16:08:51.877689939Z',
        jsonPayloadResponseCode : 'NOERROR',
        dnsQuery : {
          domain : 'cdn.utmstack.com.',
          type : 'RRSIG',
          class : 'IN',
          rValue : 'a 8 3 300 1741187631 1739286831 33820 utmstack.com. L0+ErXhgcsE5GWwFwwixqxBRKOk6yo5hdoGLX/7pIfSKlQ/OHg01HtJP+/T5k9JKXcm6eQIxo2QWMWBSDF0v4AxnM019m+a6fvYjioIS4iBC//+4s4Gm6zCPkenH2/wqUjbT4UhNjL87xdL/duRwMf6bAse71/gF+gMPQ3HOz0c=',
          ttl : '300'
        },
        jsonPayloadServerLatency : 0,
        logName : 'projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries',
        JsonPayloadAuthAnswer : true,
        insertId : '5uuba6e3n8xw',
        resourceType : 'dns_query',
        timestamp : '2025-02-13T16:08:50.651060243Z'
      },
      dataType : 'google',
      origin : {
        ip : '2400:cb00:579:1024::ac47:c924',
        domain : 'cdn.utmstack.com.',
        geolocation : {
          country : 'India',
          city : 'Mumbai',
          countryCode : 'IN',
          latitude : 19.0748,
          aso : 'CLOUDFLARENET',
          asn : 13335,
          longitude : 72.8856
        }
      },
      raw : '{"insertId":"5uuba6e3n8xw","jsonPayload":{"authAnswer":true,"destinationIP":"2001:4860:4802:36::6a","dns64Translated":false,"protocol":"UDP","queryName":"cdn.utmstack.com.","queryType":"A","responseCode":"NOERROR","serverLatency":0,"sourceIP":"2400:cb00:579:1024::ac47:c924","structuredRdata":[{"class":"IN","domainName":"cdn.utmstack.com.","rvalue":"34.144.229.76","ttl":"300","type":"A"},{"class":"IN","domainName":"cdn.utmstack.com.","rvalue":"a 8 3 300 1741187631 1739286831 33820 utmstack.com. L0+ErXhgcsE5GWwFwwixqxBRKOk6yo5hdoGLX/7pIfSKlQ/OHg01HtJP+/T5k9JKXcm6eQIxo2QWMWBSDF0v4AxnM019m+a6fvYjioIS4iBC//+4s4Gm6zCPkenH2/wqUjbT4UhNjL87xdL/duRwMf6bAse71/gF+gMPQ3HOz0c=","ttl":"300","type":"RRSIG"}]},"logName":"projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries","receiveTimestamp":"2025-02-13T16:08:51.877689939Z","resource":{"labels":{"location":"global","project_id":"utmstack-377415","source_type":"internet","target_name":"utmstack-com","target_type":"public-zone"},"type":"dns_query"},"severity":"INFO","timestamp":"2025-02-13T16:08:50.651060243Z"}',
      deviceTime : '2025-02-13T16:08:52.560801837Z',
      target : {
        ip : '2001:4860:4802:36::6a',
        geolocation : {
          country : 'United States',
          countryCode : 'US',
          latitude : 37.751,
          aso : 'GOOGLE',
          asn : 15169,
          longitude : -97.822
        }
      },
      protocol : 'UDP',
      tenantName : 'Default',
      tenantId : 'ce66672c-e36d-4761-a8c8-90058fee1a24',
      id : '903d992f-24cd-4353-abc5-33cffb27a104',
      dataSource : 'UTMStackIntegration',
      timestamp : '2025-02-13T16:08:52.560801837Z'
    } ],
    severity : 3,
    severityLabel : 'High',
    dataType : 'google',
    impact : {
      integrity : 1,
      confidentiality : 1,
      availability : 3
    },
    adversary : {
      ip : '2400:cb00:579:1024::ac47:c924',
      domain : 'cdn.utmstack.com.',
      geolocation : {
        country : 'India',
        city : 'Mumbai',
        countryCode : 'IN',
        latitude : 19.0748,
        aso : 'CLOUDFLARENET',
        asn : 13335,
        longitude : 72.8856
      }
    },
    tagRulesApplied : null,
    statusLabel : 'Open',
    target : {
      ip : '2001:4860:4802:36::6a',
      geolocation : {
        country : 'United States',
        countryCode : 'US',
        latitude : 37.751,
        aso : 'GOOGLE',
        asn : 15169,
        longitude : -97.822
      }
    },
    tags : null,
    '@timestamp' : '2025-02-13T16:08:53.486168192Z',
    statusObservation : 'This alert has been evaluated by the tag rules engine',
    name : 'Request from India',
    isIncident : false,
    category : 'Geolocation based rulees',
    impactScore : 3,
    dataSource : 'UTMStackIntegration',
    status : 2
  },
  {
    parentId: null,
    notes : '',
    description : 'This alert will report each IP from India doing requests to our CDN',
    technique : 'QA',
    reference : [ 'https://google.com' ],
    incidentDetail : {
      createdBy : '',
      observation : '',
      source : '',
      creationDate : ''
    },
    solution : '',
    id : '97a9ab4b-023d-4148-b830-4170dc221caf',
    events : [ {
      severity : 'low',
      log : {
        jsonPayloadQueryType : 'A',
        resourceLabelsLocation : 'global',
        jsonPayloadDns64Translated : false,
        resourceLabelsProjectId : 'utmstack-377415',
        receiveTimestamp : '2025-02-13T16:08:51.877689939Z',
        jsonPayloadResponseCode : 'NOERROR',
        dnsQuery : {
          domain : 'cdn.utmstack.com.',
          type : 'A',
          class : 'IN',
          rValue : '34.144.229.76',
          ttl : '300'
        },
        jsonPayloadServerLatency : 0,
        logName : 'projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries',
        JsonPayloadAuthAnswer : true,
        insertId : '5uuba6e3n8xw',
        resourceType : 'dns_query',
        timestamp : '2025-02-13T16:08:50.651060243Z'
      },
      dataType : 'google',
      origin : {
        ip : '2400:cb00:579:1024::ac47:c924',
        domain : 'cdn.utmstack.com.',
        geolocation : {
          country : 'India',
          city : 'Mumbai',
          countryCode : 'IN',
          latitude : 19.0748,
          aso : 'CLOUDFLARENET',
          asn : 13335,
          longitude : 72.8856
        }
      },
      raw : '{"insertId":"5uuba6e3n8xw","jsonPayload":{"authAnswer":true,"destinationIP":"2001:4860:4802:36::6a","dns64Translated":false,"protocol":"UDP","queryName":"cdn.utmstack.com.","queryType":"A","responseCode":"NOERROR","serverLatency":0,"sourceIP":"2400:cb00:579:1024::ac47:c924","structuredRdata":[{"class":"IN","domainName":"cdn.utmstack.com.","rvalue":"34.144.229.76","ttl":"300","type":"A"},{"class":"IN","domainName":"cdn.utmstack.com.","rvalue":"a 8 3 300 1741187631 1739286831 33820 utmstack.com. L0+ErXhgcsE5GWwFwwixqxBRKOk6yo5hdoGLX/7pIfSKlQ/OHg01HtJP+/T5k9JKXcm6eQIxo2QWMWBSDF0v4AxnM019m+a6fvYjioIS4iBC//+4s4Gm6zCPkenH2/wqUjbT4UhNjL87xdL/duRwMf6bAse71/gF+gMPQ3HOz0c=","ttl":"300","type":"RRSIG"}]},"logName":"projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries","receiveTimestamp":"2025-02-13T16:08:51.877689939Z","resource":{"labels":{"location":"global","project_id":"utmstack-377415","source_type":"internet","target_name":"utmstack-com","target_type":"public-zone"},"type":"dns_query"},"severity":"INFO","timestamp":"2025-02-13T16:08:50.651060243Z"}',
      deviceTime : '2025-02-13T16:08:52.560801837Z',
      target : {
        ip : '2001:4860:4802:36::6a',
        geolocation : {
          country : 'United States',
          countryCode : 'US',
          latitude : 37.751,
          aso : 'GOOGLE',
          asn : 15169,
          longitude : -97.822
        }
      },
      protocol : 'UDP',
      tenantName : 'Default',
      tenantId : 'ce66672c-e36d-4761-a8c8-90058fee1a24',
      id : 'b4fbf22e-7d77-41f8-b711-769e55ccebc4',
      dataSource : 'UTMStackIntegration',
      timestamp : '2025-02-13T16:08:52.560801837Z'
    } ],
    severity : 3,
    severityLabel : 'High',
    dataType : 'google',
    impact : {
      integrity : 1,
      confidentiality : 1,
      availability : 3
    },
    adversary : {
      ip : '2400:cb00:579:1024::ac47:c924',
      domain : 'cdn.utmstack.com.',
      geolocation : {
        country : 'India',
        city : 'Mumbai',
        countryCode : 'IN',
        latitude : 19.0748,
        aso : 'CLOUDFLARENET',
        asn : 13335,
        longitude : 72.8856
      }
    },
    tagRulesApplied : null,
    statusLabel : 'Open',
    target : {
      ip : '2001:4860:4802:36::6a',
      geolocation : {
        country : 'United States',
        countryCode : 'US',
        latitude : 37.751,
        aso : 'GOOGLE',
        asn : 15169,
        longitude : -97.822
      }
    },
    tags : null,
    '@timestamp' : '2025-02-13T16:08:53.461598235Z',
    statusObservation : 'This alert has been evaluated by the tag rules engine',
    name : 'Request from India',
    isIncident : false,
    category : 'Geolocation based rulees',
    impactScore : 3,
    dataSource : 'UTMStackIntegration',
    status : 2
  },
  {
    parentId: '5d1ef1c6-d725-4843-98f6-06a6b91c394e',
    notes : '',
    description : 'This alert will report each IP from India doing requests to our CDN',
    technique : 'QA',
    reference : [ 'https://google.com' ],
    incidentDetail : {
      createdBy : '',
      observation : '',
      source : '',
      creationDate : ''
    },
    solution : '',
    id : '97a9ab4b-023d-4148-b830-4170dc221caf',
    events : [ {
      severity : 'low',
      log : {
        jsonPayloadQueryType : 'A',
        resourceLabelsLocation : 'global',
        jsonPayloadDns64Translated : false,
        resourceLabelsProjectId : 'utmstack-377415',
        receiveTimestamp : '2025-02-13T16:08:51.877689939Z',
        jsonPayloadResponseCode : 'NOERROR',
        dnsQuery : {
          domain : 'cdn.utmstack.com.',
          type : 'A',
          class : 'IN',
          rValue : '34.144.229.76',
          ttl : '300'
        },
        jsonPayloadServerLatency : 0,
        logName : 'projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries',
        JsonPayloadAuthAnswer : true,
        insertId : '5uuba6e3n8xw',
        resourceType : 'dns_query',
        timestamp : '2025-02-13T16:08:50.651060243Z'
      },
      dataType : 'google',
      origin : {
        ip : '2400:cb00:579:1024::ac47:c924',
        domain : 'cdn.utmstack.com.',
        geolocation : {
          country : 'India',
          city : 'Mumbai',
          countryCode : 'IN',
          latitude : 19.0748,
          aso : 'CLOUDFLARENET',
          asn : 13335,
          longitude : 72.8856
        }
      },
      raw : '{"insertId":"5uuba6e3n8xw","jsonPayload":{"authAnswer":true,"destinationIP":"2001:4860:4802:36::6a","dns64Translated":false,"protocol":"UDP","queryName":"cdn.utmstack.com.","queryType":"A","responseCode":"NOERROR","serverLatency":0,"sourceIP":"2400:cb00:579:1024::ac47:c924","structuredRdata":[{"class":"IN","domainName":"cdn.utmstack.com.","rvalue":"34.144.229.76","ttl":"300","type":"A"},{"class":"IN","domainName":"cdn.utmstack.com.","rvalue":"a 8 3 300 1741187631 1739286831 33820 utmstack.com. L0+ErXhgcsE5GWwFwwixqxBRKOk6yo5hdoGLX/7pIfSKlQ/OHg01HtJP+/T5k9JKXcm6eQIxo2QWMWBSDF0v4AxnM019m+a6fvYjioIS4iBC//+4s4Gm6zCPkenH2/wqUjbT4UhNjL87xdL/duRwMf6bAse71/gF+gMPQ3HOz0c=","ttl":"300","type":"RRSIG"}]},"logName":"projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries","receiveTimestamp":"2025-02-13T16:08:51.877689939Z","resource":{"labels":{"location":"global","project_id":"utmstack-377415","source_type":"internet","target_name":"utmstack-com","target_type":"public-zone"},"type":"dns_query"},"severity":"INFO","timestamp":"2025-02-13T16:08:50.651060243Z"}',
      deviceTime : '2025-02-13T16:08:52.560801837Z',
      target : {
        ip : '2001:4860:4802:36::6a',
        geolocation : {
          country : 'United States',
          countryCode : 'US',
          latitude : 37.751,
          aso : 'GOOGLE',
          asn : 15169,
          longitude : -97.822
        }
      },
      protocol : 'UDP',
      tenantName : 'Default',
      tenantId : 'ce66672c-e36d-4761-a8c8-90058fee1a24',
      id : 'b4fbf22e-7d77-41f8-b711-769e55ccebc4',
      dataSource : 'UTMStackIntegration',
      timestamp : '2025-02-13T16:08:52.560801837Z'
    } ],
    severity : 3,
    severityLabel : 'High',
    dataType : 'google',
    impact : {
      integrity : 1,
      confidentiality : 1,
      availability : 3
    },
    adversary : {
      ip : '2400:cb00:579:1024::ac47:c924',
      domain : 'cdn.utmstack.com.',
      geolocation : {
        country : 'India',
        city : 'Mumbai',
        countryCode : 'IN',
        latitude : 19.0748,
        aso : 'CLOUDFLARENET',
        asn : 13335,
        longitude : 72.8856
      }
    },
    tagRulesApplied : null,
    statusLabel : 'Open',
    target : {
      ip : '2001:4860:4802:36::6a',
      geolocation : {
        country : 'United States',
        countryCode : 'US',
        latitude : 37.751,
        aso : 'GOOGLE',
        asn : 15169,
        longitude : -97.822
      }
    },
    tags : null,
    '@timestamp' : '2025-02-13T16:08:53.461598235Z',
    statusObservation : 'This alert has been evaluated by the tag rules engine',
    name : 'Request from India',
    isIncident : false,
    category : 'Geolocation based rulees',
    impactScore : 3,
    dataSource : 'UTMStackIntegration',
    status : 2
  },
  {
    parentId: '5d1ef1c6-d725-4843-98f6-06a6b91c394e',
    notes : '',
    description : 'This alert will report each IP from India doing requests to our CDN',
    technique : 'QA',
    reference : [ 'https://google.com' ],
    incidentDetail : {
      createdBy : '',
      observation : '',
      source : '',
      creationDate : ''
    },
    solution : '',
    id : 'bb820193-3a95-4c64-b9d7-ad3a228fe0c7',
    events : [ {
      severity : 'low',
      log : {
        jsonPayloadQueryType : 'DNSKEY',
        resourceLabelsLocation : 'global',
        jsonPayloadDns64Translated : false,
        resourceLabelsProjectId : 'utmstack-377415',
        receiveTimestamp : '2025-02-13T16:08:51.149380918Z',
        jsonPayloadResponseCode : 'NOERROR',
        dnsQuery : {
          domain : 'utmstack.com.',
          type : 'DNSKEY',
          class : 'IN',
          rValue : '\\# 136 010003080301000187d41ff83e7e9da01825d18a8436ee9f8aa9d75479848808ac4c5c7c93978211c5a2f4ed3d7b45fb3da0453989a7037771604cce9a517defa3803ba58b3befbf37c1344d80c8f993603a7e46d64f608f2fa402e1bbacf6fb66dfbeebd4ca78de61982c086f04e5c674c82b5fad2ff431a3e2341ded866abdb35d7f961cb6a3c1',
          ttl : '300'
        },
        jsonPayloadServerLatency : 2,
        logName : 'projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries',
        JsonPayloadAuthAnswer : true,
        insertId : 'swswf5dkee2',
        resourceType : 'dns_query',
        timestamp : '2025-02-13T16:08:50.656332005Z'
      },
      dataType : 'google',
      origin : {
        ip : '2400:cb00:579:1024::ac47:c924',
        domain : 'utmstack.com.',
        geolocation : {
          country : 'India',
          city : 'Mumbai',
          countryCode : 'IN',
          latitude : 19.0748,
          aso : 'CLOUDFLARENET',
          asn : 13335,
          longitude : 72.8856
        }
      },
      raw : '{"insertId":"swswf5dkee2","jsonPayload":{"authAnswer":true,"destinationIP":"2001:4860:4802:38::6a","dns64Translated":false,"protocol":"UDP","queryName":"utmstack.com.","queryType":"DNSKEY","responseCode":"NOERROR","serverLatency":2,"sourceIP":"2400:cb00:579:1024::ac47:c924","structuredRdata":[{"class":"IN","domainName":"utmstack.com.","rvalue":"\\\\# 264 0101030803010001dfe19fdcbf4a3ac0f5f5cd859ba39d809b5ad00becccfc692e73db66c00a926f34fbd6f03f3ab45351425f7e7c2fb7a5a7661375b5cf7fe800f1ac7f79ccd9915c24b7821972867c490f3f260ca1adb630b9423e049194555a2d4774e3f913db455a6fd97897fbf76d8fa46912aca2943c13d71481351ff099cf8d572eb4091d485d1bdb341c7bf9ca30cdbd09fad7d6e5b84bc02a397ecb538def284c09e2ea63f432022404758179a8f8a347072668fd3551b7f88547fa1ccdde3ba99c5d6b066d0ba9dcdff0d60b20e4ffb2ba36c8e462d146c2c067f3e8668cb35f6b624438d490357b7faffd7ca65e56da1757a01fe342c3049ffa7d60bb8ea63b8c676b","ttl":"300","type":"DNSKEY"},{"class":"IN","domainName":"utmstack.com.","rvalue":"\\\\# 136 010003080301000187d41ff83e7e9da01825d18a8436ee9f8aa9d75479848808ac4c5c7c93978211c5a2f4ed3d7b45fb3da0453989a7037771604cce9a517defa3803ba58b3befbf37c1344d80c8f993603a7e46d64f608f2fa402e1bbacf6fb66dfbeebd4ca78de61982c086f04e5c674c82b5fad2ff431a3e2341ded866abdb35d7f961cb6a3c1","ttl":"300","type":"DNSKEY"},{"class":"IN","domainName":"utmstack.com.","rvalue":"dnskey 8 2 300 1741187631 1739286831 58480 utmstack.com. V9fPawL7rEQbDkatGSwFhTLdR015FvwiJSOmKI6v47B2oCy4vTbkBcfjp+QgltGN1eUQbENabVbC0TTCu0Q3fdFt7gLb87FuDoCGcaNj4ZTmzX8Uj/MZd++0bLJeI662e1DyQjiSpoojKovjQnORPmKuzn+X3vvIwm9VYJ6Clz4AzcmpTNMlp9JJLJTi5NC2gKAUeOTENkWmI8ZLhRy36KQ4io+TCWLDnEN+YQTdhU6GONjGRFEsmQn/wYRiOJPA5n3+X2YCzHsvKPKqD4oxh5JpwRbSTunS7nOAIqObsBRJVyXkQyPjX3h5xYUZLmGia015vbfHg6m/dBye9lELSA==","ttl":"300","type":"RRSIG"}]},"logName":"projects/utmstack-377415/logs/dns.googleapis.com%2Fdns_queries","receiveTimestamp":"2025-02-13T16:08:51.149380918Z","resource":{"labels":{"location":"global","project_id":"utmstack-377415","source_type":"internet","target_name":"utmstack-com","target_type":"public-zone"},"type":"dns_query"},"severity":"INFO","timestamp":"2025-02-13T16:08:50.656332005Z"}',
      deviceTime : '2025-02-13T16:08:51.734589412Z',
      target : {
        ip : '2001:4860:4802:38::6a',
        geolocation : {
          country : 'United States',
          countryCode : 'US',
          latitude : 37.751,
          aso : 'GOOGLE',
          asn : 15169,
          longitude : -97.822
        }
      },
      protocol : 'UDP',
      tenantName : 'Default',
      tenantId : 'ce66672c-e36d-4761-a8c8-90058fee1a24',
      id : '42fb8682-b8b1-4fcd-b64c-a20a2447b3db',
      dataSource : 'UTMStackIntegration',
      timestamp : '2025-02-13T16:08:51.734589412Z'
    } ],
    severity : 3,
    severityLabel : 'High',
    dataType : 'google',
    impact : {
      integrity : 1,
      confidentiality : 1,
      availability : 3
    },
    adversary : {
      ip : '2400:cb00:579:1024::ac47:c924',
      domain : 'utmstack.com.',
      geolocation : {
        country : 'India',
        city : 'Mumbai',
        countryCode : 'IN',
        latitude : 19.0748,
        aso : 'CLOUDFLARENET',
        asn : 13335,
        longitude : 72.8856
      }
    },
    tagRulesApplied : null,
    statusLabel : 'Open',
    target : {
      ip : '2001:4860:4802:38::6a',
      geolocation : {
        country : 'United States',
        countryCode : 'US',
        latitude : 37.751,
        aso : 'GOOGLE',
        asn : 15169,
        longitude : -97.822
      }
    },
    tags : null,
    '@timestamp' : '2025-02-13T16:08:52.046211062Z',
    statusObservation : 'This alert has been evaluated by the tag rules engine',
    name : 'Request from India',
    isIncident : false,
    category : 'Geolocation based rulees',
    impactScore : 3,
    dataSource : 'UTMStackIntegration',
    status : 2
  }
];
