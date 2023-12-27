import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {ResizableModule} from 'angular-resizable-element';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {AssetGroupAddComponent} from './asset-group-add.component';



@NgModule({
    declarations: [
        AssetGroupAddComponent,
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
        AssetGroupAddComponent,
    ],
})
export class AssetsGroupAddModule {
}
