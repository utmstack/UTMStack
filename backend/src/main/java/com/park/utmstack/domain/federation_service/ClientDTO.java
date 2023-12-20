package com.park.utmstack.domain.federation_service;

import com.fasterxml.jackson.annotation.JsonInclude;
import org.springframework.util.StringUtils;

@JsonInclude(JsonInclude.Include.NON_ABSENT)
public class ClientDTO {
    private String clientUser;
    private String clientPassword;
    private String clientDomain;
    private String clientName;
    private String clientLogo;
    private String clientDescription;
    private String clientEmail;

    public ClientDTO() {
    }

    public ClientDTO(String str) throws Exception {
        if (!StringUtils.hasText(str))
            throw new Exception("Value of str is null or empty");

        String[] split = str.split("\\|");

        if (split.length != 3)
            throw new Exception("Decrypted value is corrupt, there has to be 4 values in the decrypted value");

        this.clientUser = split[0];
        this.clientPassword = split[1];
        this.clientName = split[2];
    }

    public ClientDTO(String clientUser, String clientPassword, String clientDomain,
                     String clientName) {
        this.clientUser = clientUser;
        this.clientPassword = clientPassword;
        this.clientDomain = clientDomain;
        this.clientName = clientName;
        this.clientLogo = null;
        this.clientDescription = null;
        this.clientEmail = null;
    }

    public String getClientUser() {
        return clientUser;
    }

    public void setClientUser(String clientUser) {
        this.clientUser = clientUser;
    }

    public String getClientPassword() {
        return clientPassword;
    }

    public void setClientPassword(String clientPassword) {
        this.clientPassword = clientPassword;
    }

    public String getClientDomain() {
        return clientDomain;
    }

    public void setClientDomain(String clientDomain) {
        this.clientDomain = clientDomain;
    }

    public String getClientName() {
        return clientName;
    }

    public void setClientName(String clientName) {
        this.clientName = clientName;
    }

    public String getClientLogo() {
        return clientLogo;
    }

    public void setClientLogo(String clientLogo) {
        this.clientLogo = clientLogo;
    }

    public String getClientDescription() {
        return clientDescription;
    }

    public void setClientDescription(String clientDescription) {
        this.clientDescription = clientDescription;
    }

    public String getClientEmail() {
        return clientEmail;
    }

    public void setClientEmail(String clientEmail) {
        this.clientEmail = clientEmail;
    }
}
