package com.park.utmstack.web.rest.compliance;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/compliance/custom")
public class CustomComplianceResource {
//    private final Logger log = LoggerFactory.getLogger(CustomComplianceResource.class);
//    private static final String CLASSNAME = "CustomComplianceResource";
//
//    private final ActiveDirectoryService activeDirectoryService;
//    private final ApplicationEventService applicationEventService;
//    private final ResponseParserFactory responseParserFactory;
//    private final UtmNetworkScanService networkScanService;
//
//    private static final Map<String, String> AD_USER_COLUMNS = new LinkedHashMap<>();
//    private static final Map<String, String> PC_COLUMNS = new LinkedHashMap<>();
//
//    static {
//        // User columns group 1
//        AD_USER_COLUMNS.put("objectSid", "Id");
//        AD_USER_COLUMNS.put("sAMAccountName", "Username");
//        AD_USER_COLUMNS.put("cn", "Common Name");
//        AD_USER_COLUMNS.put("whenCreated", "Created At");
//        AD_USER_COLUMNS.put("memberOf", "Groups");
//        AD_USER_COLUMNS.put("realLastLogon", "Last Logon");
//
//        // Computer columns
//        PC_COLUMNS.put("cn", "Name");
//        PC_COLUMNS.put("operatingSystem", "OS");
//        PC_COLUMNS.put("operatingSystemVersion", "OS Version");
//        PC_COLUMNS.put("realLastLogon", "Last Logon");
//    }
//
//    public CustomComplianceResource(ActiveDirectoryService activeDirectoryService,
//                                    ApplicationEventService applicationEventService,
//                                    ResponseParserFactory responseParserFactory,
//                                    UtmNetworkScanService networkScanService) {
//        this.activeDirectoryService = activeDirectoryService;
//        this.applicationEventService = applicationEventService;
//        this.responseParserFactory = responseParserFactory;
//        this.networkScanService = networkScanService;
//    }
//
//    /**
//     * Compliance rest endpoint for users with password never expire
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/usersWithPasswordNeverExpire")
//    public ResponseEntity<List<?>> usersWithPasswordNeverExpire() {
//        final String ctx = CLASSNAME + ".usersWithPasswordNeverExpire";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.usersWithPasswordNeverExpire();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Compliance rest endpoint for users with password expired
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/usersWithPasswordExpired")
//    public ResponseEntity<List<?>> usersWithPasswordExpired() {
//        final String ctx = CLASSNAME + ".usersWithPasswordExpired";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.usersWithPasswordExpired();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Compliance rest endpoint for get all users
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/allUsers")
//    public ResponseEntity<List<?>> allUsers() {
//        final String ctx = CLASSNAME + ".allUsers";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.allUsers();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Compliance rest endpoint for get all computers
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/allComputers")
//    public ResponseEntity<List<?>> allComputers() {
//        final String ctx = CLASSNAME + ".allComputers";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.allComputers();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, PC_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, PC_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Compliance rest endpoint for get all inactive users
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/inactiveUsers")
//    public ResponseEntity<List<?>> inactiveUsers() {
//        final String ctx = CLASSNAME + ".inactiveUsers";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.inactiveUsers();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Compliance rest endpoint for get all inactive computers
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/inactiveComputers")
//    public ResponseEntity<List<?>> inactiveComputers() {
//        final String ctx = CLASSNAME + ".inactiveComputers";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.inactiveComputers();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, PC_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, PC_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Compliance rest endpoint for get all disabled computers
//     *
//     * @return A ${@link ResponseEntity} object with a list of ${@link ActiveDirectoryInfo}
//     */
//    @GetMapping("/disabledComputers")
//    public ResponseEntity<List<?>> disabledComputers() {
//        final String ctx = CLASSNAME + ".disabledComputers";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.disabledComputers();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, PC_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, PC_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    @GetMapping("/disabledUsers")
//    public ResponseEntity<List<?>> disabledUsers() {
//        final String ctx = CLASSNAME + ".disabledUsers";
//        try {
//            Pageable page = PageRequest.of(0, 10000);
//            Optional<SearchResult> resultOpt = Optional.ofNullable(activeDirectoryService.amountOfUsersDisabledDetails(null, page));
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    @GetMapping("/enabledUsers")
//    public ResponseEntity<List<?>> enabledUsers() {
//        final String ctx = CLASSNAME + ".enabledUsers";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.enabledUsers();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    @GetMapping("/lockedOutUsers")
//    public ResponseEntity<List<?>> lockedOutUsers() {
//        final String ctx = CLASSNAME + ".lockedOutUsers";
//        try {
//            Pageable page = PageRequest.of(0, 10000);
//            Optional<SearchResult> resultOpt = Optional.ofNullable(activeDirectoryService.amountOfUsersLockoutDetails(null, page));
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    @GetMapping("/usersCreatedInLast24HByAdmins")
//    public ResponseEntity<List<?>> usersCreatedInLast24HByAdmins() {
//        final String ctx = CLASSNAME + ".usersCreatedInLast24HByAdmins";
//        try {
//            Optional<SearchResult> resultOpt = activeDirectoryService.usersCreatedInLast24HByAdmins();
//            return resultOpt.<ResponseEntity<List<?>>>map(searchResult -> ResponseEntity.ok(parseElasticResultAsListChart(searchResult, AD_USER_COLUMNS)))
//                .orElseGet(() -> ResponseEntity.ok(parseElasticResultAsListChart(null, AD_USER_COLUMNS)));
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    @GetMapping("/assetsDiscovered")
//    public ResponseEntity<List<?>> assetsDiscovered() {
//        final String ctx = CLASSNAME + ".assetsDiscovered";
//
//        try {
//            TableChartResult results = new TableChartResult();
//
//            results.addColumn("assetIp->Ip").addColumn("assetAddresses->Addresses").addColumn("assetMac->Mac")
//                .addColumn("assetOs->Os").addColumn("assetName->Name").addColumn("assetAlive->Alive").addColumn("assetStatus->Status")
//                .addColumn("assetSeverity->Severity").addColumn("discoveredAt->Discovered").addColumn("serverName->Probe");
//
//            List<UtmNetworkScan> assets = networkScanService.findAll();
//
//            if (CollectionUtils.isEmpty(assets)) {
//                results.addRow(Collections.emptyList());
//                return ResponseEntity.ok(Collections.singletonList(results));
//            }
//
//            assets.forEach(asset -> {
//                List<TableChartResult.Cell<?>> cells = new ArrayList<>();
//                cells.add(new TableChartResult.Cell<>(asset.getAssetIp(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetAddresses(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetMac(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetOs(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetName(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetAlive() ? "Yes" : "No", false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetStatus().name(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getAssetSeverity() != null ? asset.getAssetSeverity().name() : null, false));
//                cells.add(new TableChartResult.Cell<>(asset.getDiscoveredAt().toString(), false));
//                cells.add(new TableChartResult.Cell<>(asset.getServerName(), false));
//                results.addRow(cells);
//            });
//
//            return ResponseEntity.ok(Collections.singletonList(results));
//
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    /**
//     * Parse a ${@link SearchResult} object as a list of structure for chart of type list
//     *
//     * @param sr      Search result from elastic
//     * @param columns Columns to return
//     * @return A list with the structure of chart of type list
//     */
//    private List<?> parseElasticResultAsListChart(SearchResult sr, Map<String, String> columns) {
//        final String ctx = CLASSNAME + ".parseResultAsListChart";
//
//        try {
//            List<Bucket> buckets = new ArrayList<>();
//
//            columns.forEach((field, label) -> {
//                Bucket tempBucket = new Bucket();
//                tempBucket.setField(field);
//                tempBucket.setCustomLabel(label);
//                buckets.add(tempBucket);
//            });
//
//            for (int i = 0; i < buckets.size(); i++) {
//                if (i == (buckets.size() - 1))
//                    break;
//                Bucket top = buckets.get(i);
//                Bucket sub = buckets.get(i + 1);
//                top.setSubBucket(sub);
//            }
//
//            AggregationType agg = new AggregationType();
//            agg.setBucket(buckets.get(0));
//
//            UtmVisualization vis = new UtmVisualization();
//            vis.setChartType(ChartType.LIST_CHART);
//            vis.setAggregationType(agg);
//
//            ResponseParser<?> parser = responseParserFactory.instance(vis.getChartType());
//            return parser.parse(vis, sr);
//        } catch (Exception e) {
//            throw new RuntimeException(ctx + ": " + e.getMessage());
//        }
//    }
}
