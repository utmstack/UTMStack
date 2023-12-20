package com.park.utmstack.service.compliance.hipaa;

import org.springframework.stereotype.Service;

@Service
public class HipaaService {
//    private static final String CLASS_NAME = "HipaaComplianceService";
//    private final Logger log = LoggerFactory.getLogger(HipaaService.class);
//    private final JestUtil jestUtil;
//    private final ApplicationProperties applicationProperties;
//    private final ActiveDirectoryService activeDirectoryService;
//
//    public HipaaService(JestUtil jestUtil, ApplicationProperties applicationProperties,
//                        ActiveDirectoryService activeDirectoryService) {
//        this.jestUtil = jestUtil;
//        this.applicationProperties = applicationProperties;
//        this.activeDirectoryService = activeDirectoryService;
//    }
//
////    /**
////     * @return
////     * @throws UtmComplianceException
////     */
////    public List<ApplicationProperties.Tools> toolsVersionAndSignature() throws UtmComplianceException {
////        final String ctx = CLASS_NAME + ".toolsVersionAndSignature";
////        try {
////            return applicationProperties.getTools();
////        } catch (Exception e) {
////            String msg = ctx + ": " + e.getMessage();
////            log.error(msg);
////            throw new UtmComplianceException(msg);
////        }
////    }
//
//    /**
//     * @param from
//     * @param to
//     * @param top
//     * @throws UtmComplianceException
//     */
//    public List<WinlogbeatInfo> logonOrAccessAttemptsFromDeProvisionedAccounts(String from, String to, int top) throws
//        UtmComplianceException {
//        final String ctx = CLASS_NAME + ".logonOrAccessAttemptsFromDeProvisionedAccounts";
//
//        try {
//            // Getting all AD users and build his status, disabled and locked
//            Map<String, UserStatus> usersStatus = buildUserStatus(
//                activeDirectoryService.getActiveDirectoryUserInformation());
//
//            // Getting all events 4624 and 4625
//            List<WinlogbeatInfo> events = getLogonSuccessAndFailEvents(from, to, top);
//            List<WinlogbeatInfo> result = new ArrayList<>();
//
//            events.forEach(event -> {
//                String targetUserSid = event.getLogx().getWineventlog().getEventData().getTargetUserSid();
//                UserStatus usrStatus = usersStatus.get(targetUserSid);
//
//                if (usrStatus != null) {
//                    if (usrStatus.isDisabled() || usrStatus.isLocked())
//                        result.add(event);
//                } else {
//                    result.add(event);
//                }
//            });
//            return result;
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            throw new UtmComplianceException(msg);
//        }
//    }
//
//    private Map<String, UserStatus> buildUserStatus(List<ActiveDirectoryInfo> users) throws UtmComplianceException {
//        final String ctx = CLASS_NAME + ".buildUserStatus";
//        Map<String, UserStatus> result = new HashMap<>();
//
//        try {
//            users.forEach(user -> result.put(user.getObjectSid(), new UserStatus(user.getDisabled(), user.getBlocked())));
//            return result;
//        } catch (Exception e) {
//            throw new UtmComplianceException(ctx + ": " + e.getMessage());
//        }
//    }
//
//    /**
//     * @param from
//     * @param to
//     * @param top
//     * @return
//     * @throws UtmComplianceException
//     */
//    private List<WinlogbeatInfo> getLogonSuccessAndFailEvents(String from, String to, int top) throws
//        UtmComplianceException {
//        final String ctx = CLASS_NAME + ".getLogonSuccessAndFailEvents";
//        try {
//            TermsQueryBuilder winEventId = QueryBuilders.termsQuery(Constants.logxWineventlogEventId, Arrays.asList(4624, 4625));
//            RangeQueryBuilder timestampFilter = QueryBuilders.rangeQuery(Constants.timestamp).format(Constants.INDEX_TIMESTAMP_FORMAT)
//                .gte(from).lte(to);
//
//            BoolQueryBuilder bool = QueryBuilders.boolQuery().must(winEventId).must(timestampFilter);
//            SearchSourceBuilder ssb = SearchSourceBuilder.searchSource().size(top).query(bool);
//
//            SearchResult result = jestUtil.search(ssb.toString(), Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS_WINDOWS));
//
//            if (!result.isSucceeded())
//                throw new Exception(result.getErrorMessage());
//
//            if (JestUtil.getTotalHits(result) <= 0)
//                return new ArrayList<>();
//
//            return result.getHits(WinlogbeatInfo.class).stream().map(hit -> hit.source).collect(Collectors.toList());
//
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            throw new UtmComplianceException(msg);
//        }
//    }
//
//    public static class UserStatus {
//        private boolean disabled;
//        private boolean locked;
//
//        public UserStatus(boolean disabled, boolean locked) {
//            this.disabled = disabled;
//            this.locked = locked;
//        }
//
//        public boolean isDisabled() {
//            return disabled;
//        }
//
//        public void setDisabled(boolean disabled) {
//            this.disabled = disabled;
//        }
//
//        public boolean isLocked() {
//            return locked;
//        }
//
//        public void setLocked(boolean locked) {
//            this.locked = locked;
//        }
//    }


}
