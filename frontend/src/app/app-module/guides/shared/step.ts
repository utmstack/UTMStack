import {TemplateRef} from '@angular/core';

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
