import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {ResizableModule} from 'angular-resizable-element';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {AssetsApplyNoteComponent} from './assets-apply-note.component';


@NgModule({
    declarations: [
        AssetsApplyNoteComponent,
    ],
    entryComponents: [],
    imports: [
        CommonModule,
        NgbModule,
        ResizableModule,
        InfiniteScrollModule,
        FormsModule,
        NgSelectModule,
        NgxEchartsModule,
        ReactiveFormsModule,
    ],
    exports: [
        AssetsApplyNoteComponent,
    ],
})
export class AssetsApplyNoteModule {
}
