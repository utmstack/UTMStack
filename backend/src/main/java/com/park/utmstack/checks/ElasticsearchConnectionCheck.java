package com.park.utmstack.checks;

import com.park.utmstack.config.Constants;
import okhttp3.Credentials;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import java.security.cert.X509Certificate;
import java.util.Objects;

public class ElasticsearchConnectionCheck {
    private static final String CLASSNAME = "ElasticsearchConnectionCheck";

    public static ElasticsearchConnectionCheck getInstance() {
        return new ElasticsearchConnectionCheck();
    }

    public void connectionCheck(int retries) {
        final String ctx = CLASSNAME + ".connectionCheck";
        ConsoleColors.cyanBold();
        System.out.println(">> Checking elasticsearch connection:");
        do {
            try {
                pingElasticsearch();
                ConsoleColors.greenBold();
                System.out.println("\t> Success");
                ConsoleColors.reset();
                return;
            } catch (Exception e) {
                ConsoleColors.redBold();
                System.out.println("\t> Ping to elasticsearch fail: " + e.getLocalizedMessage());
                ConsoleColors.reset();
                if (retries == -1)
                    break;
                retries--;

                ConsoleColors.redBold();
                for (int i = 10; i > 0; i--) {
                    System.out.printf("\t> Retrying in: %1$s\r", i);
                    try {
                        Thread.sleep(1000L);
                    } catch (Exception ex) {
                        throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                    }
                }
            }
        } while (retries > 0);

        ConsoleColors.redBold();
        System.out.println("\t> Fail to establish connection with elasticsearch\n");
        ConsoleColors.reset();
        System.exit(1);
    }

    private void pingElasticsearch() {
        try {
            final String ELASTIC_URL = String.format("https://%1$s:%2$s",
                System.getenv(Constants.ENV_ELASTICSEARCH_HOST), System.getenv(Constants.ENV_ELASTICSEARCH_PORT));

            String user = System.getenv(Constants.ENV_ELASTICSEARCH_USER);
            String password = System.getenv(Constants.ENV_ELASTICSEARCH_PASSWORD);

            OkHttpClient client = createTrustAllClient();
            Request rq = new Request.Builder()
                .url(ELASTIC_URL)
                .header("Authorization", Credentials.basic(user, password))
                .build();
            Response rs = client.newCall(rq).execute();
            Objects.requireNonNull(rs.body()).close();
            if (!rs.isSuccessful())
                throw new RuntimeException("HTTP " + rs.code());
        } catch (Exception e) {
            throw new RuntimeException(e.getLocalizedMessage());
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
                .build();
        } catch (Exception e) {
            throw new RuntimeException("Failed to create SSL client: " + e.getMessage());
        }
    }
}
