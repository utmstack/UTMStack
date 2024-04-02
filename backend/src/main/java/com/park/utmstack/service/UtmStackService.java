package com.park.utmstack.service;

import com.park.utmstack.checks.ElasticsearchConnectionCheck;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.mail_sender.MailConfig;
import org.springframework.core.env.Environment;
import org.springframework.stereotype.Service;
import tech.jhipster.config.JHipsterConstants;

import javax.mail.MessagingException;
import javax.validation.constraints.Email;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

@Service
public class UtmStackService {
    private static final String CLASSNAME = "UtmStackService";

    private final Environment env;
    private final MailService mailService;
    private final UserService userService;

    public UtmStackService(Environment env,
                           MailService mailService,
                           UserService userService) {
        this.env = env;
        this.mailService = mailService;
        this.userService = userService;
    }

    public boolean isInDevelop() throws Exception {
        final String ctx = CLASSNAME + ".isInDevelop";
        try {
            String[] profiles = env.getActiveProfiles().length == 0 ? env.getDefaultProfiles() : env.getActiveProfiles();
            List<String> activeProfiles = Arrays.asList(profiles);
            return activeProfiles.contains(JHipsterConstants.SPRING_PROFILE_DEVELOPMENT);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public void checkEmailConfiguration(MailConfig config) throws MessagingException {
        final String ctx = CLASSNAME + ".checkEmailConfiguration";
        try {
            User user = userService.getCurrentUserLogin();
            List<String> to = Collections.singletonList(user.getEmail());
            mailService.sendCheckEmail(to, config);
        } catch (MessagingException e) {
            throw new MessagingException(ctx + ": " + e.getMessage());
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public void executeChecks() {
        ElasticsearchConnectionCheck.getInstance().connectionCheck(-1);
    }
}
