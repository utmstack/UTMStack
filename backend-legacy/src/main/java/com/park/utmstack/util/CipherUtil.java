package com.park.utmstack.util;

import javax.crypto.Cipher;
import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.PBEKeySpec;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.security.spec.KeySpec;
import java.util.Arrays;
import java.util.Base64;

public class CipherUtil {
    private static final String CLASSNAME = "CipherUtil";
    private static final String CIPHER_INSTANCE = "AES/CBC/PKCS5Padding";
    private static final int ITERATION_COUNT = 65536;
    private static final int KEY_LENGTH = 128;

    private static class CryptoContext {
        final SecretKeySpec secretKey;
        final IvParameterSpec iv;

        CryptoContext(SecretKeySpec secretKey, IvParameterSpec iv) {
            this.secretKey = secretKey;
            this.iv = iv;
        }
    }

    private static CryptoContext getCryptoContext(String myKey) throws Exception {
        final String ctx = CLASSNAME + ".getCryptoContext";
        try {
            byte[] salt = myKey.getBytes(StandardCharsets.UTF_8);
            MessageDigest sha = MessageDigest.getInstance("SHA-1");
            salt = sha.digest(salt);
            KeySpec spec = new PBEKeySpec(myKey.toCharArray(), salt, ITERATION_COUNT, KEY_LENGTH);
            SecretKeyFactory f = SecretKeyFactory.getInstance("PBKDF2WithHmacSHA1");
            byte[] key = f.generateSecret(spec).getEncoded();
            SecretKeySpec secretKey = new SecretKeySpec(key, "AES");
            IvParameterSpec iv = new IvParameterSpec(Arrays.copyOf(salt, 16));
            return new CryptoContext(secretKey, iv);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public static String encrypt(String str, String secret) {
        final String ctx = CLASSNAME + ".encrypt";
        try {
            CryptoContext context = getCryptoContext(secret);
            Cipher cipher = Cipher.getInstance(CIPHER_INSTANCE);
            cipher.init(Cipher.ENCRYPT_MODE, context.secretKey, context.iv);
            return Base64.getEncoder().encodeToString(cipher.doFinal(str.getBytes(StandardCharsets.UTF_8)));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public static String decrypt(String str, String secret) {
        final String ctx = CLASSNAME + ".decrypt";
        try {
            CryptoContext context = getCryptoContext(secret);
            Cipher cipher = Cipher.getInstance(CIPHER_INSTANCE);
            cipher.init(Cipher.DECRYPT_MODE, context.secretKey, context.iv);
            return new String(cipher.doFinal(Base64.getDecoder().decode(str)));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public static String generateSafeToken() {
        SecureRandom random = new SecureRandom();
        return Base64.getEncoder().encodeToString(random.generateSeed(64));
    }
}
