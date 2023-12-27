export interface MultilineResponse {
  categories: string[];
  series: { serie: string, values: number[] }[];
}

export interface SerieValue {
  serie: string;
  value: number;
}
