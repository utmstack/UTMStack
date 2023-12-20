package com.park.utmstack.service.application_modules;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.enums.ModuleRequirementStatus;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import org.springframework.stereotype.Service;

@Service
public class ModuleRequirementChecksService {
    private static final String CLASSNAME = "ModuleRequirementChecksService";

    private final ElasticsearchService elasticsearchService;

    public ModuleRequirementChecksService(ElasticsearchService elasticsearchService) {
        this.elasticsearchService = elasticsearchService;
    }

    /**
     * Check if exist windows event logs
     *
     * @return A ${@link ModuleRequirement} with requirement check result
     * @throws Exception In case of any error
     */
    public ModuleRequirement checkWindowsEvents() throws Exception {
        final String ctx = CLASSNAME + ".checkWindowsEvents";
        try {
            boolean indexExist = elasticsearchService.indexExist(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS_WINDOWS));

            ModuleRequirement requirement = new ModuleRequirement();
            requirement.setCheckName(Constants.MODULE_CHECK_WINEVENTLOG);
            requirement.setCheckStatus(ModuleRequirementStatus.OK);

            if (!indexExist) {
                requirement.setCheckStatus(ModuleRequirementStatus.FAIL);
                requirement.setFailMessage("Windows events are not being logged");
            }
            return requirement;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
