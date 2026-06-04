type Callback = (data: any) => void;

declare class WebfluxEventSource {

  onmessage: Callback;
  onerror: Callback;
  close: () => void;

  constructor(name: string);

  addEventListener(event: string, cb: Callback): void;
}
