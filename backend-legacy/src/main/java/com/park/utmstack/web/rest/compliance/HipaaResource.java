package com.park.utmstack.web.rest.compliance;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/compliance/hipaa")
public class HipaaResource {
//    private final Logger log = LoggerFactory.getLogger(HipaaResource.class);
//    private static final String CLASS_NAME = "HipaaResource";
//
//    private final HipaaService hipaaService;
//    private final ApplicationEventService applicationEventService;
//
//    public HipaaResource(HipaaService hipaaService,
//                         ApplicationEventService applicationEventService) {
//        this.hipaaService = hipaaService;
//        this.applicationEventService = applicationEventService;
//    }
//
//    @GetMapping("/logon-or-access-attempts-from-deprovisioned-accounts")
//    public ResponseEntity<List<WinlogbeatInfo>> getLogonOrAccessAttemptsFromDeProvisionedAccounts(@RequestParam String from,
//                                                                                                  @RequestParam String to,
//                                                                                                  @RequestParam Integer top) {
//        final String ctx = CLASS_NAME + ".getLogonOrAccessAttemptsFromDeProvisionedAccounts";
//        try {
//            List<WinlogbeatInfo> result = hipaaService.logonOrAccessAttemptsFromDeProvisionedAccounts(from, to, top);
//            return ResponseEntity.ok(result);
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    @PostMapping("/logon-or-access-attempts-from-deprovisioned-accounts/csv")
//    public ResponseEntity<Void> getLogonOrAccessAttemptsFromDeProvisionedAccountsToCsv(
//        @RequestBody CsvExportingParams params, HttpServletResponse response) {
//        final String ctx = CLASS_NAME + ".getLogonOrAccessAttemptsFromDeProvisionedAccountsToCsv";
//        try {
//            Map<String, String> paramMap = buildParameters(params.getFilters());
//
//            List<WinlogbeatInfo> result = hipaaService.logonOrAccessAttemptsFromDeProvisionedAccounts(paramMap.get("from"),
//                paramMap.get("to"),
//                params.getTop());
//            UtilCsv.prepareToDownload(response, params.getColumns(), result);
//            return ResponseEntity.ok().build();
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
//                HeaderUtil.createFailureAlert("", "", msg)).body(null);
//        }
//    }
//
//    private Map<String, String> buildParameters(List<FilterType> filters) {
//        Map<String, String> result = new HashMap<>();
//        for (FilterType filter : filters) {
//            if (filter.getField().equals("@timestamp")) {
//                List value = (List) filter.getValue();
//                result.put("from", (String) value.get(0));
//                result.put("to", (String) value.get(1));
//            }
//        }
//        return result;
//    }
}
