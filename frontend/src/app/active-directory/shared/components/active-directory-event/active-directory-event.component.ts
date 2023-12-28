import {Component, Input, OnInit} from '@angular/core';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {TreeObjectBehavior} from '../../behavior/tree-object.behvior';
import {WinlogbeatEventType} from '../../types/winlogbeat-event.type';

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
  event: WinlogbeatEventType;

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

  onEventChange($event: WinlogbeatEventType) {
    this.event = $event;
    this.message = this.event ? this.replaceDetail($event.logx.wineventlog.message) : '';
  }
}
