import {TemplateRef} from '@angular/core';

export const COLLECTOR_MESSAGE = 'Enable log collector.<br>' +
  'Follow the instructions below based on your operating system and preferred protocol to enable the log collector.\n';

export enum ContentType {
  COMMAND
}

export interface Image {
  alt: string;
  src: string;
}
export interface Content {
  id: string;
  type?: string;
  commands?: string[];
  images?: Image[];
}
export interface Step {
  id: string;
  name: string;
  content?: Content;
}
