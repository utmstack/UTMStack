import {TemplateRef} from '@angular/core';

export enum ContentType {
  COMMAND
}
export interface Content {
  id: string;
  type?: string;
  commands?: string[];
}
export interface Step {
  id: string;
  name: string;
  content?: Content;
}
