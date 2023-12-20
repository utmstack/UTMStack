package com.park.utmstack.service.licence;

import com.park.utmstack.domain.UtmClient;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.licence.Licence;
import com.park.utmstack.domain.licence.LicenceResponse;
import com.park.utmstack.domain.licence.LicenceStatus;
import com.park.utmstack.service.UtmClientService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.util.exceptions.UtmInvalidLicenceException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Objects;

@Service
public class LicenceManagerService {
    private static final String CLASSNAME = "LicenceManagerService";
    private final Logger log = LoggerFactory.getLogger(LicenceManagerService.class);

    private static final String CONSUMER_KEY = "ck_ad89e93e4d5ea9435c5736cd5c7a2ebcf0e26f3c";
    private static final String CONSUMER_SECRET = "cs_c5fd29d6a453bb0d3f318620e53a00ea1fb555aa";

    private static final String GET_LICENCE_URL = "https://utmstack.com/wp-json/lmfwc/v2/licenses/%1$s?consumer_key=%2$s&consumer_secret=%3$s";
    private static final String ACTIVATE_LICENCE_URL = "https://utmstack.com/wp-json/lmfwc/v2/licenses/activate/%1$s?consumer_key=%2$s&consumer_secret=%3$s";

    private final RestTemplateService httpUtil;
    private final UtmClientService clientService;
    private final ApplicationEventService applicationEventService;

    public LicenceManagerService(RestTemplateService httpUtil, UtmClientService clientService,
                                 ApplicationEventService applicationEventService) {
        this.httpUtil = httpUtil;
        this.clientService = clientService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * Obtains the information of a license and validates that it can be activated.
     * It is considered that a license can be activated when:
     * 1. The licence status is Delivered
     * 2. It's not expired
     * 3. It has never been activated (timesActivated is null or <= 0)
     *
     * @param licenceKey Licence key to validate
     * @return {@code true} if the licence key is valid
     */
    private boolean canActivateLicense(String licenceKey) throws Exception {
        final String ctx = CLASSNAME + ".canActivateLicense";
        try {
            // Requesting licence information from licence manager server
            LicenceResponse licenceResponse = getLicence(licenceKey);
            Licence licence = Objects.requireNonNull(licenceResponse, String.format("Licence with key: %1$s not found", licenceKey)).getData();

            // Checking condition for a valid licence
            return licence.getStatus() == LicenceStatus.DELIVERED.getStatusValue()
                && licence.getExpiresAtAsInstant().isAfter(LocalDateTime.now().toInstant(ZoneOffset.UTC))
                && (Objects.isNull(licence.getTimesActivated()) || licence.getTimesActivated() + 1 <= licence.getTimesActivatedMax());
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Check if license is valid
     *
     * @param licence Key of the licence to validate
     * @return {@code true} if the licence key is valid
     */
    private boolean isValidLicense(Licence licence) {
        return licence.getStatus() == LicenceStatus.DELIVERED.getStatusValue()
            && licence.getExpiresAtAsInstant().isAfter(LocalDateTime.now().toInstant(ZoneOffset.UTC));
    }


    /**
     * Gets a licence by his key
     *
     * @param licenceKey Licence key
     * @return A response entity of {@link LicenceResponse}
     * @throws Exception In case of any error
     */
    private LicenceResponse getLicence(String licenceKey) throws Exception {
        final String ctx = CLASSNAME + ".getLicence";
        final String url = String.format(GET_LICENCE_URL, licenceKey, CONSUMER_KEY, CONSUMER_SECRET);
        try {
            ResponseEntity<LicenceResponse> rs = httpUtil.get(url, LicenceResponse.class);
            if (!rs.getStatusCode().is2xxSuccessful())
                throw new Exception(httpUtil.extractErrorMessage(rs));
            return rs.getBody();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Activate a licence, adds 1 to the times activated attribute
     *
     * @param licenceKey Key of the licence to activate
     * @return A response entity of {@link LicenceResponse}
     * @throws Exception In case of any error
     */
    public void activateLicence(String licenceKey, String email) throws Exception {
        final String ctx = CLASSNAME + ".activateLicence";
        final String url = String.format(ACTIVATE_LICENCE_URL, licenceKey, CONSUMER_KEY, CONSUMER_SECRET);
        try {
            // Checking if a licence is valid before activate it
            if (!canActivateLicense(licenceKey))
                throw new UtmInvalidLicenceException(String.format("Licence %1$s can't be activated because is invalid", licenceKey));

            // Activating licence
            ResponseEntity<LicenceResponse> rs = httpUtil.get(url, LicenceResponse.class);
            if (!rs.getStatusCode().is2xxSuccessful())
                throw new Exception(httpUtil.extractErrorMessage(rs));

            // License information after being activated
            Licence licence = Objects.requireNonNull(rs.getBody()).getData();

            // Getting the client information
            List<UtmClient> clients = clientService.findAll();
            UtmClient client = !CollectionUtils.isEmpty(clients) ? clients.get(0) : new UtmClient();

            // Save or update client information
            client.setClientMail(email);
            client.setClientLicenceId(licenceKey);
            client.setClientLicenceVerified(true);
            client.setClientLicenceCreation(licence.getCreatedAtAsInstant());
            client.setClientLicenceExpire(licence.getExpiresAtAsInstant());

            clientService.save(client);
        } catch (UtmInvalidLicenceException e) {
            throw new UtmInvalidLicenceException(ctx + ": " + e.getMessage());
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Scheduled(fixedDelay = 1800000, initialDelay = 60000)
    public void checkLicence() {
        final String ctx = CLASSNAME + ".checkLicence";

        try {
            List<UtmClient> clients = clientService.findAll();

            // Exit if the client information was not found
            if (CollectionUtils.isEmpty(clients))
                return;

            UtmClient client = clients.get(0);

            // Exit if the client has not a licence to check
            if (!StringUtils.hasText(client.getClientLicenceId()))
                return;

            // Requesting licence information from licence manager server
            LicenceResponse licenceResponse = getLicence(client.getClientLicenceId());
            Licence licence = Objects.requireNonNull(licenceResponse).getData();

            if (!isValidLicense(licence)) {
                client.setClientLicenceVerified(false);
                client.setClientLicenceId(null);
                client.setClientLicenceCreation(null);
                client.setClientLicenceExpire(null);

                clientService.save(client);
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    public boolean checkClientLicence() throws Exception {
        final String ctx = CLASSNAME + ".checkClientLicence";

        try {
            // Getting the client information
            List<UtmClient> clients = clientService.findAll();

            // Validating client information
            if (CollectionUtils.isEmpty(clients))
                throw new Exception("Client information not found");

            UtmClient client = clients.get(0);

            if (!StringUtils.hasText(client.getClientLicenceId()))
                return false;

            // Requesting licence information from licence manager server
            LicenceResponse rs = getLicence(client.getClientLicenceId());
            Licence licence = Objects.requireNonNull(rs).getData();

            // Checking if the licence is valid
            boolean validLicense = isValidLicense(licence);

            client.setClientLicenceVerified(validLicense);

            if (!validLicense) {
                client.setClientLicenceId(null);
                client.setClientLicenceCreation(null);
                client.setClientLicenceExpire(null);
            }

            clientService.save(client);
            return validLicense;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return false;
        }
    }
}
