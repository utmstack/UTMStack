package com.park.utmstack.service.soc_ai;

import com.google.gson.Gson;
import com.park.utmstack.domain.UtmAlertSocaiProcessingRequest;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.service.UtmAlertSocaiProcessingRequestService;
import com.park.utmstack.service.application_modules.UtmModuleService;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Service
public class SocAIService {
    private static final String CLASSNAME = "SocAIService";
    private final Logger log = LoggerFactory.getLogger(SocAIService.class);
    private final String SOCAI_PROCESS_URL;

    private final UtmAlertSocaiProcessingRequestService socaiProcessingRequestService;
    private final UtmModuleService moduleService;

    public SocAIService(UtmAlertSocaiProcessingRequestService socaiProcessingRequestService,
                        UtmModuleService moduleService) {
        this.socaiProcessingRequestService = socaiProcessingRequestService;
        this.moduleService = moduleService;
        SOCAI_PROCESS_URL = System.getenv("SOC_AI_BASE_URL");
    }


    public void sendData(Object data) {
        final String ctx = CLASSNAME + ".sendData";
        try {
            OkHttpClient client = new OkHttpClient.Builder()
                .connectTimeout(10, TimeUnit.SECONDS)
                .writeTimeout(10, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .build();
            MediaType mediaType = MediaType.parse("application/json; charset=utf-8");
            RequestBody body = RequestBody.create(new Gson().toJson(data), mediaType);
            Request request = new Request.Builder().url(SOCAI_PROCESS_URL).post(body)
                .addHeader("Content-Type", "application/json").build();

            try (Response rs = client.newCall(request).execute()) {
                if (!rs.isSuccessful())
                    throw new Exception(ctx + "Unexpected response: " + rs);
            }
        } catch (Exception e) {
            log.error(ctx + ": " + e.getLocalizedMessage());
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    @Scheduled(fixedDelay = 30000)
    public void sendRequests() {
        final String ctx = CLASSNAME + ".sendRequests";
        try {
            if (!moduleService.isModuleActive(ModuleName.SOC_AI))
                return;

            if (!StringUtils.hasText(SOCAI_PROCESS_URL)) {
                log.error(ctx + ": Environment variable SOC_AI_BASE_URL is missing or does not have a value");
                return;
            }

            Page<UtmAlertSocaiProcessingRequest> requests = socaiProcessingRequestService.findAll(PageRequest.of(0, 20));
            if (!requests.hasContent())
                return;
            List<String> ids = requests.getContent().stream().map(UtmAlertSocaiProcessingRequest::getAlertId)
                .collect(Collectors.toList());
            try {
                sendData(ids);
                socaiProcessingRequestService.delete(ids);
            } catch (Exception e) {
                log.error(ctx + ": " + e.getLocalizedMessage());
            }
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    @Async
    public void requestSocAiProcess(List<String> alertIds) {
        final String ctx = CLASSNAME + ".requestSocAiProcess";
        try {
            List<UtmAlertSocaiProcessingRequest> ids = alertIds.stream().map(UtmAlertSocaiProcessingRequest::new)
                .collect(Collectors.toList());
            socaiProcessingRequestService.saveAll(ids);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
