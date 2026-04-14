package com.park.utmstack.service.soc_ai;

import com.google.gson.Gson;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.shared_types.alert.UtmAlert;
import com.park.utmstack.service.application_modules.UtmModuleService;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import java.security.cert.X509Certificate;
import java.util.concurrent.TimeUnit;

@Service
public class SocAIService {
    private static final String CLASSNAME = "SocAIService";
    private final Logger log = LoggerFactory.getLogger(SocAIService.class);
    private final String SOCAI_BASE_URL;
    private final String SOCAI_ANALYZE_ENDPOINT = "/api/v1/analyze";
    private final OkHttpClient httpClient;

    private final UtmModuleService moduleService;

    public SocAIService(UtmModuleService moduleService) {
        this.moduleService = moduleService;
        SOCAI_BASE_URL = System.getenv("SOC_AI_BASE_URL");
        this.httpClient = createTrustAllClient();
    }

    /**
     * Sends a complete alert to SOC-AI for analysis
     */
    public void analyzeAlert(UtmAlert alert) {
        final String ctx = CLASSNAME + ".analyzeAlert";
        try {
            if (!StringUtils.hasText(SOCAI_BASE_URL)) {
                throw new RuntimeException("SOC_AI_BASE_URL environment variable is not configured");
            }

            String internalKey = System.getenv(Constants.ENV_INTERNAL_KEY);
            if (!StringUtils.hasText(internalKey)) {
                throw new RuntimeException("INTERNAL_KEY environment variable is not configured");
            }

            MediaType mediaType = MediaType.parse("application/json; charset=utf-8");
            RequestBody body = RequestBody.create(new Gson().toJson(alert), mediaType);

            String url = SOCAI_BASE_URL + SOCAI_ANALYZE_ENDPOINT;
            Request request = new Request.Builder()
                .url(url)
                .post(body)
                .addHeader("Content-Type", "application/json")
                .addHeader("X-Internal-Key", internalKey)
                .build();

            try (Response rs = httpClient.newCall(request).execute()) {
                if (!rs.isSuccessful()) {
                    String responseBody = rs.body() != null ? rs.body().string() : "No response body";
                    throw new Exception("Unexpected response: " + rs.code() + " - " + responseBody);
                }
            }
        } catch (Exception e) {
            log.error(ctx + ": " + e.getLocalizedMessage());
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private OkHttpClient createTrustAllClient() {
        try {
            TrustManager[] trustAllCerts = new TrustManager[]{
                new X509TrustManager() {
                    public X509Certificate[] getAcceptedIssuers() { return new X509Certificate[0]; }
                    public void checkClientTrusted(X509Certificate[] certs, String authType) {}
                    public void checkServerTrusted(X509Certificate[] certs, String authType) {}
                }
            };

            SSLContext sslContext = SSLContext.getInstance("TLS");
            sslContext.init(null, trustAllCerts, new java.security.SecureRandom());

            return new OkHttpClient.Builder()
                .sslSocketFactory(sslContext.getSocketFactory(), (X509TrustManager) trustAllCerts[0])
                .hostnameVerifier((hostname, session) -> true)
                .connectTimeout(10, TimeUnit.SECONDS)
                .writeTimeout(10, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .build();
        } catch (Exception e) {
            throw new RuntimeException("Failed to create SSL client: " + e.getMessage());
        }
    }
}
