package com.park.utmstack.service.federation_service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.federation_service.UtmFederationServiceClient;
import com.park.utmstack.repository.federation_service.UtmFederationServiceClientRepository;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.util.RandomUtil;
import com.park.utmstack.util.CipherUtil;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Base64Utils;

/**
 * Service Implementation for managing UtmFederationServiceClient.
 */
@Service
@Transactional
public class UtmFederationServiceClientService {

    private static final String CLASSNAME = "UtmFederationServiceClientService";

    private final UtmFederationServiceClientRepository federationServiceClientRepository;
    private final UserService userService;

    public UtmFederationServiceClientService(UtmFederationServiceClientRepository federationServiceClientRepository,
                                             UserService userService) {
        this.federationServiceClientRepository = federationServiceClientRepository;
        this.userService = userService;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASSNAME + ".init";
        try {
            if (federationServiceClientRepository.findById(1L).isEmpty())
                generateApiToken();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Activate the federation service integration system
     *
     * @return A string with a token configuration
     */
    public String generateApiToken() {
        final String ctx = CLASSNAME + ".generateApiToken";
        try {
            String instanceName = System.getenv(Constants.ENV_SERVER_NAME);
            String fsPassword = RandomUtil.generatePassword();

            userService.createFederationServiceUser(fsPassword);

            String token = String.join("|", instanceName, CipherUtil.encrypt(fsPassword, System.getenv(Constants.ENV_ENCRYPTION_KEY)));
            token = Base64Utils.encodeToUrlSafeString(token.getBytes());

            UtmFederationServiceClient fsClient = new UtmFederationServiceClient();
            fsClient.setId(1L);
            fsClient.setFsClientToken(token);
            federationServiceClientRepository.save(fsClient);

            return token;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get the connection key
     *
     * @return The connection key
     */
    public String getFederationServiceManagerToken() {
        final String ctx = CLASSNAME + ".getFederationServiceManagerToken";
        try {
            return federationServiceClientRepository.findById(1L)
                .map(UtmFederationServiceClient::getFsClientToken).orElse(null);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Search for the token and if it was found a match return true otherwise false
     *
     * @param token Token to search for
     * @return True if the token exist, false otherwise
     */
    public boolean existToken(String token) {
        final String ctx = CLASSNAME + ".existToken";
        try {
            return federationServiceClientRepository.findByFsClientToken(token).isPresent();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
