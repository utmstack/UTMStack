import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {TranslateModule} from '@ngx-translate/core';
import {InlineSVGModule} from 'ng-inline-svg';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {IncidentResponseSharedModule} from '../../../incident-response/shared/incident-response-shared.module';
import {UtmSharedModule} from '../../../shared/utm-shared.module';
import {DataMgmtSharedModule} from '../../data-mgmt-shared/data-mgmt-shared.module';

import {AlertRulesComponent} from '../alert-rules/alert-rules.component';
import {AlertTagsCreateComponent} from '../alert-tags/alert-tags-create/alert-tags-create.component';
import {AlertTagsManagementComponent} from '../alert-tags/alert-tags-management.component';
import {AlertTagsRenderComponent} from '../alert-tags/alert-tags-render/alert-tags-render.component';
import {AlertTagsViewComponent} from '../alert-tags/alert-tags-view/alert-tags-view.component';
import {AlertApplyIncidentComponent} from './components/alert-actions/alert-apply-incident/alert-apply-incident.component';
import {AlertApplyNoteComponent} from './components/alert-actions/alert-apply-note/alert-apply-note.component';
import {AlertApplyStatusComponent} from './components/alert-actions/alert-apply-status/alert-apply-status.component';
import {AlertTagsApplyComponent} from './components/alert-actions/alert-apply-tags/alert-tags-apply.component';
import {
  AlertSeverityDescriptionComponent
} from './components/alert-actions/alert-severity-description/alert-severity-description.component';
import {AlertCategoryComponent} from './components/alert-category/alert-category.component';
import {AlertCompleteComponent} from './components/alert-complete/alert-complete.component';
import {AlertDescriptionComponent} from './components/alert-description/alert-description.component';
import {AlertDocUpdateInProgressComponent} from './components/alert-doc-update-in-progress/alert-doc-update-in-progress.component';
import {AlertFullLogComponent} from './components/alert-full-log/alert-full-log.component';
import {AlertHistoryComponent} from './components/alert-history/alert-history.component';
import {AlertHostDetailComponent} from './components/alert-host-detail/alert-host-detail.component';
import {AlertIncidentDetailComponent} from './components/alert-incident-detail/alert-incident-detail.component';
import {AlertIpComponent} from './components/alert-ip/alert-ip.component';
import {AlertMapLocationComponent} from './components/alert-map-location/alert-map-location.component';
import {AlertProposedSolutionComponent} from './components/alert-proposed-solution/alert-proposed-solution.component';
import {AlertRuleCreateComponent} from './components/alert-rule-create/alert-rule-create.component';
import {AlertRuleDetailComponent} from './components/alert-rule-detail/alert-rule-detail.component';
import {AlertRuleTagRelatedComponent} from './components/alert-rule-tag-related/alert-rule-tag-related.component';
import {AlertSeverityComponent} from './components/alert-severity/alert-severity.component';
import {AlertSocAiComponent} from './components/alert-soc-ai/alert-soc-ai.component';
import {AlertStatusViewComponent} from './components/alert-status-view/alert-status-view.component';
import {AlertStatusComponent} from './components/alert-status/alert-status.component';
import {AlertTagLabelComponent} from './components/alert-tag-label/alert-tag-label.component';
import {AlertViewDetailComponent} from './components/alert-view-detail/alert-view-detail.component';
import {AlertViewLastChangeComponent} from './components/alert-view-last-change/alert-view-last-change.component';
import {DataFieldRenderComponent} from './components/data-field-render/data-field-render.component';
import {ActiveFiltersComponent} from './components/filters/active-filters/active-filters.component';
import {AlertFilterComponent} from './components/filters/alert-filter/alert-filter.component';
import {AlertGenericFilterComponent} from './components/filters/alert-generic-filter/alert-generic-filter.component';
import {FilterAppliedComponent} from './components/filters/filter-applied/filter-applied.component';
import {RowToFiltersComponent} from './components/filters/row-to-filter/row-to-filters.component';
import {StatusFilterComponent} from './components/filters/status-filter/status-filter.component';

@NgModule({
  declarations: [
    AlertStatusComponent,
    AlertSeverityComponent,
    AlertCompleteComponent,
    AlertFilterComponent,
    ActiveFiltersComponent,
    RowToFiltersComponent,
    AlertGenericFilterComponent,
    StatusFilterComponent,
    DataFieldRenderComponent,
    AlertTagsApplyComponent,
    AlertTagsViewComponent,
    AlertTagsManagementComponent,
    AlertTagsViewComponent,
    AlertTagsCreateComponent,
    AlertApplyStatusComponent,
    AlertTagsRenderComponent,
    FilterAppliedComponent,
    AlertApplyNoteComponent,
    AlertIpComponent,
    AlertHistoryComponent,
    AlertMapLocationComponent,
    AlertSeverityDescriptionComponent,
    AlertFullLogComponent,
    AlertHostDetailComponent,
    AlertApplyIncidentComponent,
    AlertViewLastChangeComponent,
    AlertRulesComponent,
    AlertDescriptionComponent,
    AlertCategoryComponent,
    AlertProposedSolutionComponent,
    AlertDocUpdateInProgressComponent,
    AlertRuleCreateComponent,
    AlertRuleDetailComponent,
    AlertTagLabelComponent,
    AlertRuleTagRelatedComponent,
    AlertViewDetailComponent,
    AlertStatusViewComponent,
    AlertIncidentDetailComponent,
    AlertSocAiComponent
  ],
  entryComponents: [
    AlertStatusComponent,
    AlertSeverityComponent,
    AlertCompleteComponent,
    ActiveFiltersComponent,
    RowToFiltersComponent,
    AlertTagsManagementComponent,
    AlertTagsCreateComponent,
    AlertGenericFilterComponent,
    AlertTagsApplyComponent,
    ActiveFiltersComponent,
    AlertRuleCreateComponent],
  exports: [
    AlertStatusComponent,
    AlertSeverityComponent,
    AlertCompleteComponent,
    ActiveFiltersComponent,
    RowToFiltersComponent,
    AlertGenericFilterComponent,
    AlertFilterComponent,
    DataFieldRenderComponent,
    StatusFilterComponent,
    AlertTagsCreateComponent,
    AlertTagsApplyComponent,
    AlertTagsManagementComponent,
    AlertApplyStatusComponent,
    AlertTagsRenderComponent,
    FilterAppliedComponent,
    AlertApplyNoteComponent,
    AlertHistoryComponent,
    AlertMapLocationComponent,
    AlertSeverityDescriptionComponent,
    AlertFullLogComponent,
    AlertHostDetailComponent,
    AlertApplyIncidentComponent,
    AlertViewLastChangeComponent,
    AlertRulesComponent,
    AlertDescriptionComponent,
    AlertCategoryComponent,
    AlertProposedSolutionComponent,
    AlertRuleCreateComponent,
    AlertRuleDetailComponent,
    AlertTagLabelComponent,
    AlertRuleTagRelatedComponent,
    AlertViewDetailComponent,
    AlertStatusViewComponent,
    AlertIncidentDetailComponent,
    AlertSocAiComponent
  ],
  imports: [
    CommonModule,
    TranslateModule,
    UtmSharedModule,
    FormsModule,
    NgbModule,
    ReactiveFormsModule,
    NgSelectModule,
    InfiniteScrollModule,
    NgxJsonViewerModule,
    IncidentResponseSharedModule,
    DataMgmtSharedModule,
    RouterModule,
    InlineSVGModule,
  ]
})
export class AlertManagementSharedModule {
}
