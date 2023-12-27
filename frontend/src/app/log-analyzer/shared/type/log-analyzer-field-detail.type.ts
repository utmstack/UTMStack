export class LogAnalyzerFieldDetailType {
  total?: number;
  top?: LogAnalyzerFieldDetailTopType[];
}

export class LogAnalyzerFieldDetailTopType {
  count?: number;
  percent?: number;
  value: string;
}

export class LogAnalyzerChartResponseType {
  categories: string[];
  values: number[];
}
