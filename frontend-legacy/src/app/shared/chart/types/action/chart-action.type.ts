export class ChartActionType {
  active?: boolean;
  customUrl?: boolean;
  navigate?: string;
  customNavigate?: string;

  constructor(active?: boolean,
              customUrl?: boolean,
              navigate?: string,
              customNavigate?: string) {
    this.active = active;
    this.customUrl = customUrl;
    this.navigate = navigate ? navigate : null;
    this.customNavigate = customNavigate ? customNavigate : null;
  }
}
