import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {FederationInstance} from '../domain/federation-instance.model';

const STORAGE_KEY = 'utm.federation.activeInstanceId';

@Injectable({providedIn: 'root'})
export class FederationInstanceStateService {
  private instancesSubject = new BehaviorSubject<FederationInstance[]>([]);
  private activeSubject = new BehaviorSubject<FederationInstance | null>(null);
  private switchSubject = new BehaviorSubject<number>(0);

  readonly instances$: Observable<FederationInstance[]> = this.instancesSubject.asObservable();
  readonly active$: Observable<FederationInstance | null> = this.activeSubject.asObservable();
  readonly switch$: Observable<number> = this.switchSubject.asObservable();

  setInstances(instances: FederationInstance[]): void {
    this.instancesSubject.next(instances);
    if (instances.length === 0) {
      this.activeSubject.next(null);
      return;
    }
    const persistedId = this.readPersistedId();
    const fromStorage = persistedId !== null ? instances.find(i => i.id === persistedId) : undefined;
    const current = this.activeSubject.getValue();
    const stillThere = current ? instances.find(i => i.id === current.id) : undefined;
    const next = stillThere || fromStorage || instances[0];
    this.setActive(next, false);
  }

  setActive(instance: FederationInstance, broadcastSwitch: boolean = true): void {
    const current = this.activeSubject.getValue();
    if (current && current.id === instance.id) {
      return;
    }
    this.activeSubject.next(instance);
    this.persistId(instance.id);
    if (broadcastSwitch) {
      this.switchSubject.next(this.switchSubject.getValue() + 1);
    }
  }

  clear(): void {
    this.instancesSubject.next([]);
    this.activeSubject.next(null);
    this.clearPersistedId();
  }

  get current(): FederationInstance | null {
    return this.activeSubject.getValue();
  }

  get instances(): FederationInstance[] {
    return this.instancesSubject.getValue();
  }

  private readPersistedId(): number | null {
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) {
        return null;
      }
      const parsed = Number(raw);
      return Number.isFinite(parsed) ? parsed : null;
    } catch {
      return null;
    }
  }

  private persistId(id: number): void {
    try {
      window.localStorage.setItem(STORAGE_KEY, String(id));
    } catch {
      // ignore
    }
  }

  private clearPersistedId(): void {
    try {
      window.localStorage.removeItem(STORAGE_KEY);
    } catch {
      // ignore
    }
  }
}
