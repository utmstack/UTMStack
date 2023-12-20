export class UtmIndexPattern {
  id: number;
  pattern: string;
  patternSystem?: boolean;
  patternModule?: string;

  constructor(id: number, pattern: string, patternSystem: boolean, patternModule?: string) {
    this.id = id;
    this.pattern = pattern;
    this.patternSystem = patternSystem;
    this.patternModule = patternModule;
  }
}
