import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Subject} from 'rxjs';
import {debounceTime, distinctUntilChanged, takeUntil} from 'rxjs/operators';

@Component({
  selector: 'app-team-members-search',
  templateUrl: './team-members-search.component.html'
})
export class TeamMembersSearchComponent implements OnInit, OnDestroy {
  @Input() placeholder = 'Search by name, login or email';
  @Output() searchChange = new EventEmitter<string>();

  value = '';
  private input$ = new Subject<string>();
  private destroy$ = new Subject<void>();

  ngOnInit(): void {
    this.input$
      .pipe(debounceTime(300), distinctUntilChanged(), takeUntil(this.destroy$))
      .subscribe(term => this.searchChange.emit(term));
  }

  onInput(value: string): void {
    this.value = value;
    this.input$.next(value);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
