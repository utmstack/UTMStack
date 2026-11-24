import {Component, EventEmitter, Input, OnChanges, Output, SimpleChanges} from '@angular/core';

export interface AllowlistEntry {
  value: string;
  kind: 'IP' | 'CIDR' | 'Domain';
}

const IPV4_REGEX = /^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}$/;
const IPV6_REGEX = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|::(ffff(:0{1,4})?:)?((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d))$/;
const DOMAIN_LABEL_REGEX = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

@Component({
  standalone: false,
  selector: 'app-utm-allowlist',
  templateUrl: './utm-allowlist.component.html',
  styleUrls: ['./utm-allowlist.component.scss']
})
export class UtmAllowlistComponent implements OnChanges {
  @Input() value = '';
  @Input() disabled = false;
  @Output() valueChange = new EventEmitter<string>();

  entries: AllowlistEntry[] = [];
  newEntry = '';
  filterText = '';
  errorMessage = '';

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['value']) {
      this.entries = this.parseRawValue(this.value || '');
    }
  }

  get filteredEntries(): AllowlistEntry[] {
    const q = this.filterText.trim().toLowerCase();
    if (!q) {
      return this.entries;
    }
    return this.entries.filter(
      e => e.value.toLowerCase().includes(q) || e.kind.toLowerCase().includes(q)
    );
  }

  addEntries(): void {
    this.errorMessage = '';
    const raw = (this.newEntry || '').trim();
    if (!raw) {
      return;
    }

    const tokens = raw.split(/[,\r\n]+/).map(t => t.trim()).filter(t => t.length > 0);
    const existing = new Set(this.entries.map(e => e.value.toLowerCase()));
    const toAdd: AllowlistEntry[] = [];

    for (const token of tokens) {
      const validated = this.validateAndClassify(token);
      if (!validated) {
        this.errorMessage = `Invalid entry "${token}". Enter a valid IPv4/IPv6 address, CIDR range (e.g. 203.0.113.0/24), or domain (e.g. example.com or *.example.com).`;
        return;
      }
      const key = validated.value.toLowerCase();
      if (!existing.has(key)) {
        existing.add(key);
        toAdd.push(validated);
      }
    }

    if (toAdd.length > 0) {
      this.entries = [...this.entries, ...toAdd];
      this.emitChange();
    }
    this.newEntry = '';
  }

  onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter') {
      event.preventDefault();
      this.addEntries();
    }
  }

  removeEntry(entryValue: string): void {
    this.errorMessage = '';
    this.entries = this.entries.filter(e => e.value !== entryValue);
    this.emitChange();
  }

  clearAll(): void {
    this.errorMessage = '';
    this.entries = [];
    this.filterText = '';
    this.emitChange();
  }

  private emitChange(): void {
    const serialized = this.entries.map(e => e.value).join(',');
    this.valueChange.emit(serialized);
  }

  private parseRawValue(raw: string): AllowlistEntry[] {
    if (!raw || !raw.trim()) {
      return [];
    }
    const tokens = raw.split(/[,\r\n]+/).map(t => t.trim()).filter(t => t.length > 0);
    const seen = new Set<string>();
    const result: AllowlistEntry[] = [];
    for (const token of tokens) {
      const classified = this.validateAndClassify(token);
      if (!classified) {
        continue;
      }
      const key = classified.value.toLowerCase();
      if (seen.has(key)) {
        continue;
      }
      seen.add(key);
      result.push(classified);
    }
    return result;
  }

  private validateAndClassify(raw: string): AllowlistEntry | null {
    const trimmed = raw.trim();
    if (!trimmed) {
      return null;
    }

    if (trimmed.includes('/')) {
      const parts = trimmed.split('/');
      if (parts.length !== 2 || !parts[0] || !/^\d+$/.test(parts[1])) {
        return null;
      }
      const prefix = Number(parts[1]);
      if (IPV4_REGEX.test(parts[0]) && prefix >= 0 && prefix <= 32) {
        return {value: `${parts[0]}/${prefix}`, kind: 'CIDR'};
      }
      if (IPV6_REGEX.test(parts[0]) && prefix >= 0 && prefix <= 128) {
        return {value: `${parts[0].toLowerCase()}/${prefix}`, kind: 'CIDR'};
      }
      return null;
    }

    if (IPV4_REGEX.test(trimmed)) {
      return {value: trimmed, kind: 'IP'};
    }

    if (trimmed.includes(':')) {
      if (IPV6_REGEX.test(trimmed)) {
        return {value: trimmed.toLowerCase(), kind: 'IP'};
      }
      return null;
    }

    let lower = trimmed.toLowerCase();
    if (lower.endsWith('.')) {
      lower = lower.slice(0, -1);
    }
    const wildcard = lower.startsWith('*.');
    const host = wildcard ? lower.slice(2) : lower;
    if (!host || host.length > 253) {
      return null;
    }
    if (wildcard && !host.includes('.')) {
      return null;
    }
    const labels = host.split('.');
    if (labels.every(l => /^\d+$/.test(l))) {
      return null;
    }
    if (!labels.every(l => DOMAIN_LABEL_REGEX.test(l))) {
      return null;
    }
    return {value: wildcard ? `*.${host}` : host, kind: 'Domain'};
  }
}
