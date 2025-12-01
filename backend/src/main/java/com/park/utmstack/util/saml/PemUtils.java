package com.park.utmstack.util.saml;

import com.park.utmstack.util.exceptions.FileProcessingException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.multipart.MultipartFile;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.KeyFactory;
import java.security.PrivateKey;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.security.spec.PKCS8EncodedKeySpec;
import java.util.Base64;

@Slf4j
public class PemUtils {

    private static final String className = PemUtils.class.getName();

    public static PrivateKey parsePrivateKey(String pemContent) {
        try {
            String base64 = pemContent
                    .replace("-----BEGIN PRIVATE KEY-----", "")
                    .replace("-----END PRIVATE KEY-----", "")
                    .replaceAll("\\s+", "");
            byte[] decoded = Base64.getDecoder().decode(base64);

            PKCS8EncodedKeySpec keySpec = new PKCS8EncodedKeySpec(decoded);
            KeyFactory keyFactory = KeyFactory.getInstance("RSA");
            return keyFactory.generatePrivate(keySpec);
        } catch (Exception e) {
            log.error("{} Error parsing PEM private key", className + ": "+ "parsePrivateKey" , e);
            throw new IllegalArgumentException("Failed to parse PEM private key", e);
        }
    }

    /**
     * Parse a PEM string into an X509Certificate.
     *
     * @param pemContent certificate in PEM format
     * @return X509Certificate
     */
    public static X509Certificate parseCertificate(String pemContent) {
        try {
            String base64 = pemContent
                    .replace("-----BEGIN CERTIFICATE-----", "")
                    .replace("-----END CERTIFICATE-----", "")
                    .replaceAll("\\s+", "");
            byte[] decoded = Base64.getDecoder().decode(base64);
            CertificateFactory factory = CertificateFactory.getInstance("X.509");
            return (X509Certificate) factory.generateCertificate(new ByteArrayInputStream(decoded));
        } catch (Exception e) {
            log.error("{} Error parsing PEM certificate", className + ": "+ "parseCertificate", e);
            throw new IllegalArgumentException("Failed to parse PEM certificate");
        }
    }

    public static void validateFilesForCreate(MultipartFile privateKeyFile, MultipartFile certificateFile) {
        if (privateKeyFile == null || privateKeyFile.isEmpty()) {
            throw new FileProcessingException("The private key is required");
        }
        if (certificateFile == null || certificateFile.isEmpty()) {
            throw new IllegalArgumentException("The certificate is required");
        }
        validateFileContent(privateKeyFile, certificateFile);
    }

    public static void validateFilesForUpdate(MultipartFile privateKeyFile, MultipartFile certificateFile) {
        boolean hasPrivateKey = privateKeyFile != null && !privateKeyFile.isEmpty();
        boolean hasCertificate = certificateFile != null && !certificateFile.isEmpty();

        if (hasPrivateKey && !hasCertificate) {
            throw new FileProcessingException("The certificate is required");
        }
        if (hasCertificate && !hasPrivateKey) {
            throw new FileProcessingException("The private key is required");
        }
        if (hasPrivateKey) {
            validateFileContent(privateKeyFile, certificateFile);
        }
    }


    private static void validateFileContent(MultipartFile privateKeyFile, MultipartFile certificateFile) {
        if (privateKeyFile != null && !privateKeyFile.isEmpty()) {
            try {
                String content = new String(privateKeyFile.getBytes(), StandardCharsets.UTF_8);
                PemUtils.parsePrivateKey(content);
            } catch (IOException e) {
                throw new FileProcessingException("An error occurred while reading the private key file");
            } catch (Exception e) {
                throw new FileProcessingException("The PEM private key is invalid");
            }
        }

        if (certificateFile != null && !certificateFile.isEmpty()) {
            try {
                String content = new String(certificateFile.getBytes(), StandardCharsets.UTF_8);
                PemUtils.parseCertificate(content);
            } catch (IOException e) {
                throw new IllegalArgumentException("An error occurred while reading the certificate file");
            } catch (Exception e) {
                throw new FileProcessingException("The PEM certificate is invalid");
            }
        }
    }
}


