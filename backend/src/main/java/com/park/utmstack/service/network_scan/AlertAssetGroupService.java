package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.repository.network_scan.UtmAssetGroupRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
@RequiredArgsConstructor
@Slf4j
public class AlertAssetGroupService {

    private static final String CLASSNAME = "AlertAssetGroupService";

    private final UtmNetworkScanRepository networkScanRepository;


    @Transactional(readOnly = true)
    public Map<String, Map<String, Object>> getAssetGroupsMapForAlerts() {
        final String ctx = CLASSNAME + ".getAssetGroupsMapForAlerts";

        try {
            List<Object[]> results = networkScanRepository.findAllAssetGroupMappings();

            Map<String, Map<String, Object>> assetGroupsMap = new HashMap<>();

            for (Object[] row : results) {
                String assetName = (String) row[0];
                Long groupId = (Long) row[1];
                String groupName = (String) row[2];

                Map<String, Object> groupInfo = new HashMap<>();
                groupInfo.put("id", groupId);
                groupInfo.put("name", groupName);

                assetGroupsMap.put(assetName, groupInfo);
            }

            return assetGroupsMap;

        } catch (Exception e) {
            log.error("{}: Error retrieving asset groups map: {}", ctx, e.getMessage());
            return new HashMap<>();
        }
    }
}