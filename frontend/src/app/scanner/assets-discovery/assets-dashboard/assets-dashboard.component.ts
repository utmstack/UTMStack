import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {VsSeverityResolverService} from '../../../shared/services/scan/vs-severity-resolver.service';
import {AssetSeverityHelpComponent} from '../../shared/components/asset-severity-help/asset-severity-help.component';
import {AssetModel} from '../../shared/model/assets/asset.model';
import {BarHostDef} from '../shared/chart/bar-host.def';
import {HostTopologyDef} from '../shared/chart/host-topology.def';
import {MultilineSeverityDef} from '../shared/chart/multiline-severity.def';
import {PieSeverityClassDef} from '../shared/chart/pie-severity-class.def';
import {AssetsService} from '../shared/services/assets.service';
import {AssetDashboardService} from './shared/services/asset-dashboard.service';

@Component({
  selector: 'app-assets-dashboard',
  templateUrl: './assets-dashboard.component.html',
  styleUrls: ['./assets-dashboard.component.scss']
})
export class AssetsDashboardComponent implements OnInit {
  multilineOption: any;
  pieOption: any;
  barHostVulnerabilitiesOption: any;
  barSoVulnerabilitiesOption: any;
  topologyOption: any;
  pieOsOption: any;
  pdfExport = false;
  loadingPieOption = true;
  loadingLineOption = true;
  loadingBarHostVulnerabilitiesOption = true;
  loadingBarSoVulnerabilitiesOption = true;
  loadingPieOsOption = true;
  loadingTopologyOption: any;

  constructor(private assetDashboardService: AssetDashboardService,
              private severityResolver: VsSeverityResolverService,
              private assetService: AssetsService,
              private pieSeverityClassDef: PieSeverityClassDef,
              private hostTopologyDef: HostTopologyDef,
              private barHostDef: BarHostDef,
              private multilineSeverityDef: MultilineSeverityDef,
              private router: Router,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    window.addEventListener('beforeprint', (event) => {
      this.pdfExport = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.pdfExport = false;
    });
    this.getHostsByModificationTime();
    this.getHostsBySeverityClass();
    this.getOperatingSystemsBySeverityClass();
    this.getAssets();
    this.getAssetsOs();
  }

  exportToPdf() {
    this.pdfExport = true;
    // captureScreen('assetDashboard').then((finish) => {
    //   this.pdfExport = false;
    // });
    setTimeout(() => {
      window.print();
    }, 1000);
  }

  getAssets() {
    this.assetService.query({type: 'host', size: 500000}).subscribe(assets => {
      this.getHostsTopology(assets.body);
      this.getMostVulnerableHost(assets.body);
    });
  }

  getAssetsOs() {
    this.assetService.query({type: 'os', size: 500000}).subscribe(assets => {
      this.getOperatingSystemsByVulnerabilityScore(assets.body);
    });
  }

  // multiline chart
  getHostsByModificationTime() {
    this.loadingLineOption = true;
    this.assetDashboardService.hostsByModificationTime().subscribe(value => {
      this.loadingLineOption = false;
      this.multilineOption = this.multilineSeverityDef.buildChartHostsByModificationTime(value.body);
    });
  }

  // severity chart
  getHostsBySeverityClass() {
    this.loadingPieOption = true;
    this.assetDashboardService.hostsBySeverityClass().subscribe(value => {
      this.loadingPieOption = false;
      if (value.body !== null) {
        this.pieOption = this.pieSeverityClassDef.buildChartBySeverityClass(value.body[0]);
      }
    });
  }

  getHostsTopology(assets: AssetModel[]) {
    this.loadingTopologyOption = false;
    this.topologyOption = this.hostTopologyDef.buildChartHostTopologyDef(assets);
  }

  //  end severity chart
  getMostVulnerableHost(assets: AssetModel[]) {
    this.loadingBarHostVulnerabilitiesOption = true;
    this.assetDashboardService.mostVulnerableHost().subscribe(value => {
      this.barHostVulnerabilitiesOption = this.barHostDef.buildCharMostVulnerableHost(value.body[0], assets);
      this.loadingBarHostVulnerabilitiesOption = false;
    });
  }

  getOperatingSystemsByVulnerabilityScore(assets: AssetModel[]) {
    this.loadingBarSoVulnerabilitiesOption = true;
    this.assetDashboardService.operatingSystemsByVulnerabilityScore().subscribe(value => {
      this.loadingBarSoVulnerabilitiesOption = false;
      this.barSoVulnerabilitiesOption = this.barHostDef.buildCharOperatingSystemsByVulnerabilityScore(value.body[0], assets);
    });
  }

  getOperatingSystemsBySeverityClass() {
    this.loadingPieOsOption = true;
    this.assetDashboardService.operatingSystemsBySeverityClass().subscribe(value => {
      if (value.body !== null) {
        this.pieOsOption = this.pieSeverityClassDef.buildChartOperatingSystemsBySeverityClass(value.body[0]);
        this.loadingPieOsOption = false;
      }
    });
  }

  navigateFilteredByStatus($event: any) {
    this.router.navigate(['/scanner/assets-discovery/assets'],
      {queryParams: {severity: $event.name}});
  }

  navigateFilteredByHost($event: any) {
    if (!$event.name.includes('>')) {
      this.router.navigate(['/scanner/assets-discovery/assets'],
        {queryParams: {host: $event.name}});
    }
  }

  navigateFilteredByHostSo($event: any) {
    this.router.navigate(['/scanner/assets-discovery/assets'],
      {queryParams: {hostSo: $event.name}});
  }

  viewSeverityHelp() {
    const modal = this.modalService.open(AssetSeverityHelpComponent, {centered: true});
  }
}
