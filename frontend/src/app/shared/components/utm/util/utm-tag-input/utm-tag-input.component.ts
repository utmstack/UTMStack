import {Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {find} from 'lodash';

@Component({
  selector: 'app-utm-tag-input',
  templateUrl: './utm-tag-input.component.html',
  styleUrls: ['./utm-tag-input.component.scss']
})
export class UtmTagInputComponent implements OnInit {
  @ViewChild('tagInput') tagInputRef: ElementRef;
  @Input() tags: string[] = [];
  @Output() tagsChange = new EventEmitter<string[]>();
  form: FormGroup;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.form = this.fb.group({
      tag: [undefined],
    });
  }

  focusTagInput(): void {
    this.tagInputRef.nativeElement.focus();
  }

  onKeyUp(event: KeyboardEvent): void {
    const inputValue: string = this.form.controls.tag.value;
    if (event.code === 'Enter' && inputValue) {
      this.addTag(inputValue);
      this.form.controls.tag.setValue('');
    }
  }

  addTag(tag: string): void {
    if (tag[tag.length - 1] === ',' || tag[tag.length - 1] === ' ') {
      tag = tag.slice(0, -1);
    }
    if (tag.length > 0 && !find(this.tags, tag)) {
      this.tags.push(tag);
    }
    this.tagsChange.emit(this.tags);
  }

  removeTag(tag?: string): void {
    const index = this.tags.indexOf(tag);
    this.tags.splice(index, 1);
    if (this.tags.length === 0) {
      this.tagsChange.emit(null);
    } else {
      this.tagsChange.emit(this.tags);
    }
  }

}
