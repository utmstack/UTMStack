import '@testing-library/jest-dom'

// jsdom has no layout engine and no ResizeObserver; components that measure
// their container only need the observer to exist.
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver ??= ResizeObserverStub
