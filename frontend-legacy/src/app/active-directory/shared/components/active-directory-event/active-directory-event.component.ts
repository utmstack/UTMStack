import {Component, Input, OnInit} from '@angular/core';
import {Event} from '../../../../shared/types/event/event';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {TreeObjectBehavior} from '../../behavior/tree-object.behvior';

@Component({
  selector: 'app-active-directory-event',
  templateUrl: './active-directory-event.component.html',
  styleUrls: ['./active-directory-event.component.scss']
})
export class AdEventComponent implements OnInit {
  @Input() objectId: any;
  @Input() eventsFilter: string[];
  @Input() time: TimeFilterType;
  message: string;
  event: Event;

  constructor(private treeObjectBehavior: TreeObjectBehavior) {
  }

  ngOnInit() {
    // this.treeObjectBehavior.$objectId.next(this.objectId);
  }

  replaceDetail(message: string): string {
    let msg = message.split('\n').join('<br>');
    msg = String(msg).split('\t\t').join('&nbsp;');
    msg = String(msg).split('\t').join('&nbsp;&nbsp;');
    return msg;
  }

  onEventChange($event: Event) {
    this.event = $event;
    this.message = this.event ? this.replaceDetail(this.event.log.message) : '';
  }
}
