package com.park.utmstack.web.rest.vm;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.Getter;
import lombok.Setter;

import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.util.HashMap;
import java.util.Map;

/**
 * View Model object for storing a user's credentials.
 */
@Setter
public class LoginVM implements AuditableDTO {

    @Getter
    @NotNull
    @Size(min = 1, max = 50)
    private String username;

    @Getter
    @NotNull
    @Size(min = ManagedUserVM.PASSWORD_MIN_LENGTH, max = ManagedUserVM.PASSWORD_MAX_LENGTH)
    private String password;

    private Boolean rememberMe;

    public Boolean isRememberMe() {
        return rememberMe;
    }

    @Override
    public String toString() {
        return "LoginVM{" +
            "username='" + username + '\'' +
            ", rememberMe=" + rememberMe +
            '}';
    }

    @Override
    public Map<String, Object> toAuditMap() {
        Map<String, Object> context = new HashMap<>();

        context.put("loginAttempt", this.username);

        return context;
    }
}
