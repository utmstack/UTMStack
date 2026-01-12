
export class PlanPrice {
  plan_id: string;
  id: string;
  price: number;
  currency: string;
  interval: 'day' | 'week' | 'month' | 'year';
}



export class PlanModel {
  id: string;
  name: string;
  description: string;
  default_price_id: string;
  active_projects: number;
  ai_requests: number;
  allowed_users: number;
  custom_domain: boolean;
  position: number;
  max_resources: number;
  project_level_roles: boolean;
  support: boolean;
  visual_editor: boolean
  prices: PlanPrice[];
}



export class SubscriptionModel {
  price_id: string|undefined;
  name: string|undefined;
  active_projects: number|undefined;
  ai_requests: number|undefined;
  allowed_users: number|undefined;
  custom_domain: boolean|undefined;
  position: number|undefined;
  max_resources: number|undefined;
  project_level_roles: boolean|undefined;
  quantity: number|undefined;
  status: string|undefined;
  current_period_start: string|undefined;
  current_period_end: string|undefined;
  cancel_at_period_end: boolean|undefined;
  support: boolean|undefined;
  visual_editor: boolean
  from_promotion_code?:boolean
}

export class StripeUrlModel {
  url: string;
}
