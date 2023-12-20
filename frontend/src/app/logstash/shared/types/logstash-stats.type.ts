export class UtmLogstashPipeline {
  general: PipelineGeneral;
  pipelines: UtmPipeline[];
}

export class PipelineGeneral {
  host: string;
  version: string;
  status: string;
  pipeline: GeneralStat;
  jvm: GeneralJvm;
}

export class GeneralStat {
  workers: number;
  batchSize: number;
  batchDelay: number;
}

export class GeneralJvm {
  threads: GeneralThreads;
  mem: GeneralMem;
  uptimeInMillis: number;
}

export class GeneralThreads {
  count: number;
  peakCount: number;
}

export class GeneralMem {
  heapUsedPercent: number;
  heapCommittedInBytes: number;
  heapMaxInBytes: number;
  heapUsedInBytes: number;
  nonHeapUsedInBytes: number;
  nonHeapCommittedInBytes: number;
}

export class UtmPipeline {
  id: number;
  pipelineId: string;
  pipelineName: string;
  parentPipeline?: any; // Use the appropriate type here
  pipelineStatus: string;
  moduleName: string;
  systemOwner: boolean;
  events: PipelineEvents;
  reloads: PipelineReloads;
}

export class PipelineEvents {
  in: number;
  filtered: number;
  out: number;
}

export class PipelineReloads {
  lastFailureTimestamp?: string; // Use the appropriate type here
  successes: number;
  failures: number;
  lastError?: PipelineLastError; // Use the appropriate type here
  lastSuccessTimestamp?: string; // Use the appropriate type here
}

export class PipelineLastError {
  message?: string; // Use the appropriate type here
}
