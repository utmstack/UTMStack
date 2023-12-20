package com.park.utmstack.service.licence;

import com.park.utmstack.domain.UtmClient;
import com.park.utmstack.domain.licence.LicenceType;
import com.park.utmstack.service.UtmClientService;
import com.park.utmstack.service.WebClientService;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.List;
import java.util.Objects;

@Service
public class LicenceService {
    private static final String CLASSNAME = "LicenceService";

    private static final String LICENCE_BASE_URL = "https://ls.utmstack.com";

    private final UtmClientService clientService;
    private final WebClientService webClientService;

    public LicenceService(UtmClientService clientService, WebClientService webClientService) {
        this.clientService = clientService;
        this.webClientService = webClientService;
    }

    /**
     * This method take a licence string as parameter and send a request to
     * https://ls.utmstack.com/check/{licence} to check if the licence is valid
     *
     * @param licence Licence key to check
     * @throws Exception In case of any error
     */
    public LicenceType checkLicence(String licence) throws Exception {
        final String ctx = CLASSNAME + ".checkLicence";
        try {
            if (!StringUtils.hasText(licence))
                throw new Exception("Parameter [licence] is missing");

            final String LICENCE_CHECK_URI = "/check/" + licence;

            LicenceType licenceDetails = webClientService.webClient(LICENCE_BASE_URL).
                get().uri(LICENCE_CHECK_URI).retrieve().bodyToMono(LicenceType.class).block();

            Objects.requireNonNull(licenceDetails, "Licence check service response is null");

            updateClientInformation(licenceDetails);

            return licenceDetails;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * @param name
     * @param email
     * @param licence
     * @return
     * @throws Exception
     */
    public LicenceType.Status activateLicence(String name, String email, String licence) throws Exception {
        final String ctx = CLASSNAME + ".activateLicence";

        try {
            Assert.hasText(name, "Missing value of param [name]");
            Assert.hasText(email, "Missing value of param [email]");
            Assert.hasText(licence, "Missing value of param [licence]");

            final String LICENCE_ACTIVATE_URI = String.format("/activate/%1$s/%2$s/%3$s", licence, name, email);

            LicenceType licenceDetails = webClientService.webClient(LICENCE_BASE_URL).
                get().uri(LICENCE_ACTIVATE_URI).retrieve().bodyToMono(LicenceType.class).block();

            Objects.requireNonNull(licenceDetails, "Licence activate service response is null");

            updateClientInformation(licenceDetails);

            return LicenceType.Status.status(licenceDetails.getDetails().getStatus());

        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }


    }

    /**
     * Update client information with the licence check service response
     *
     * @param licenceDetails : Licence information
     * @throws Exception In case of any error
     */
    private void updateClientInformation(LicenceType licenceDetails) throws Exception {
        final String ctx = CLASSNAME + ".updateClientInformation";

        try {
            Objects.requireNonNull(licenceDetails, "Parameter [licenceDetails] is null");

            List<UtmClient> clients = clientService.findAll();

            if (CollectionUtils.isEmpty(clients))
                throw new Exception("Client not found");

            UtmClient client = clients.get(0);
            client.setClientLicenceVerified(licenceDetails.getDetails().getStatus() == LicenceType.Status.VALID.status());
            client.setClientLicenceId(licenceDetails.getDetails().getKey());
            client.setClientLicenceCreation(licenceDetails.getDetails().getActivation());
            client.setClientLicenceExpire(licenceDetails.getExpire());

            clientService.save(client);

        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }


}
