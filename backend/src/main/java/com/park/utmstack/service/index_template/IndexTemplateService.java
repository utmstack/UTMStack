package com.park.utmstack.service.index_template;

import com.park.utmstack.config.ApplicationProperties;
import com.park.utmstack.domain.index_template.CreateTemplateResponse;
import com.park.utmstack.domain.index_template.IndexTemplate;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

@Service
public class IndexTemplateService {
    private static final String CLASSNAME = "IndexTemplateService";
    private final Logger log = LoggerFactory.getLogger(IndexTemplateService.class);

    private final RestTemplateService httpUtil;
    private String ELASTIC_URL;

    public IndexTemplateService(RestTemplateService httpUtil,
                                ApplicationProperties properties) {
        this.httpUtil = httpUtil;
        //ELASTIC_URL = properties.getChartBuilder().getElasticsearchUrl();
    }

    /**
     * To check if a specific template exists
     *
     * @param templateName: Template name
     * @return true if template exist, false otherwise
     */
    public boolean existTemplate(String templateName) throws Exception {
        final String ctx = CLASSNAME + "existTemplate";
        final String EXIST_TPL_URL = String.format("%1$s/_index_template/%2$s", ELASTIC_URL, templateName);
        try {
            ResponseEntity<String> tplExist = httpUtil.head(EXIST_TPL_URL);
            return tplExist.getStatusCode().equals(HttpStatus.OK);
        } catch (Exception e) {
            String msg = ctx + ".existTemplate";
            log.error(msg);
            throw new Exception(e.getMessage());
        }
    }

    /**
     * Creates a template and applies it to any new index whose name matches the pattern.
     *
     * @param templateName  Name of the template
     * @param policyId      Identifier of the policy to apply
     * @param indexPatterns List of index patterns to the new template will be applied
     * @throws Exception In case of any error
     */
    public void createIndexTemplateWithPolicy(String templateName, String policyId, String... indexPatterns) throws Exception {
        final String ctx = CLASSNAME + "createIndexTemplateWithPolicy";
        final String CREATE_TPL_URL = String.format("%1$s/_index_template/%2$s", ELASTIC_URL, templateName);

        try {
            IndexTemplate indexTemplate = IndexTemplate.builder()
                .withIndexPattern(indexPatterns)
                .withPriority(1)
                .withIndexStateManagementPolicyIdSetting(policyId)
                .withNumberOfShards(3)
                .withNumberOfReplicas(0)
                .withTotalFieldsLimit(50000)
                .build();
            httpUtil.put(CREATE_TPL_URL, indexTemplate, CreateTemplateResponse.class);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new Exception(msg);
        }
    }
}
