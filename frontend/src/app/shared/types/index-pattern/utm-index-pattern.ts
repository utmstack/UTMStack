export class UtmIndexPattern {
  id: number;
  pattern: string;
  patternSystem?: boolean;
  patternModule?: string;
  active?: boolean;

  constructor(id: number, pattern: string, patternSystem: boolean, patternModule?: string, active?: boolean) {
    this.id = id;
    this.pattern = pattern;
    this.patternSystem = patternSystem;
    this.patternModule = patternModule;
    this.active = active;
  }
}
