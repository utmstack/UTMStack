package com.park.utmstack.web.rest;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.PasswordChangeDTO;
import com.park.utmstack.service.dto.UserDTO;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.web.rest.errors.*;
import com.park.utmstack.web.rest.vm.KeyAndPasswordVM;
import com.park.utmstack.web.rest.vm.ManagedUserVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import javax.validation.Valid;
import java.util.Optional;

/**
 * REST controller for managing the current user's account.
 */
@RestController
@RequestMapping("/api")
public class AccountResource {

    private static final String CLASSNAME = "AccountResource";

    private final Logger log = LoggerFactory.getLogger(AccountResource.class);

    private final UserRepository userRepository;

    private final UserService userService;

    private final MailService mailService;

    private final ApplicationEventService applicationEventService;

    public AccountResource(UserRepository userRepository, UserService userService, MailService mailService, ApplicationEventService applicationEventService) {

        this.userRepository = userRepository;
        this.userService = userService;
        this.mailService = mailService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * {@code GET  /authenticate} : check if the user is authenticated, and return its login.
     *
     * @param request the HTTP request
     * @return the login if the user is authenticated
     */
    @GetMapping("/authenticate")
    public String isAuthenticated(HttpServletRequest request) {
        log.debug("REST request to check if the current user is authenticated");
        return request.getRemoteUser();
    }

    /**
     * {@code GET  /account} : get the current user.
     *
     * @return the current user
     * @throws RuntimeException {@code 500 (Internal Server Error)} if the user couldn't be returned
     */
    @GetMapping("/account")
    public UserDTO getAccount() {
        return userService.getUserWithAuthorities()
            .map(UserDTO::new)
            .orElseThrow(() -> new InternalServerErrorException("User could not be found"));
    }

    /**
     * {@code POST  /account} : update the current user information.
     *
     * @param userDTO the current user information
     * @throws EmailAlreadyUsedException {@code 400 (Bad Request)} if the email is already used
     * @throws RuntimeException {@code 500 (Internal Server Error)} if the user login wasn't found
     */
    @PostMapping("/account")
    public void saveAccount(@Valid @RequestBody UserDTO userDTO) {
        String userLogin = SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new InternalServerErrorException("Current user login not found"));
        Optional<User> existingUser = userRepository.findOneByEmailIgnoreCase(userDTO.getEmail());
        if (existingUser.isPresent() && (!existingUser.get().getLogin().equalsIgnoreCase(userLogin))) {
            throw new EmailAlreadyUsedException();
        }
        Optional<User> user = userRepository.findOneByLogin(userLogin);
        if (!user.isPresent()) {
            throw new InternalServerErrorException("User could not be found");
        }
        userService.updateUser(userDTO.getFirstName(), userDTO.getLastName(), userDTO.getEmail(),
            userDTO.getLangKey(), userDTO.getImageUrl());
    }

    /**
     * {@code POST  /account/change-password} : changes the current user's password.
     *
     * @param passwordChangeDto current and new password
     * @throws InvalidPasswordException {@code 400 (Bad Request)} if the new password is incorrect
     */
    @PostMapping(path = "/account/change-password")
    @AuditEvent(
        attemptType = ApplicationEventType.PASSWORD_CHANGE_ATTEMPT,
        attemptMessage = "Attempting to change current user's password",
        successType = ApplicationEventType.PASSWORD_CHANGE_SUCCESS,
        successMessage = "User's password changed successfully"
    )
    public void changePassword(@RequestBody PasswordChangeDTO passwordChangeDto) {
        final String ctx = CLASSNAME + ".changePassword";
        try {
            validatePasswordLength(passwordChangeDto.getNewPassword());
            userService.changePassword(passwordChangeDto.getCurrentPassword(), passwordChangeDto.getNewPassword());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * {@code POST   /account/reset-password/init} : Send an Email to reset the password of the user.
     * <p>
     * Always returns 200 OK, whether or not the submitted email is registered, so that
     * account existence cannot be inferred from the response (anti-enumeration).
     *
     * @param mail the mail of the user
     */
    @PostMapping(path = "/account/reset-password/init")
    public void requestPasswordReset(@RequestBody String mail) {
        final String ctx = CLASSNAME + ".requestPasswordReset";
        try {
            if (!StringUtils.hasText(mail))
                return;

            Optional<User> user = userService.requestPasswordReset(mail);

            if (user.isPresent()) {
                try {
                    mailService.sendPasswordResetMail(user.get());
                } catch (Exception e) {
                    // The mail was not sent, but the response must stay uniform: logging
                    // only, no distinct error surfaced to the caller.
                    String msg = ctx + ": Failed to send password reset mail: " + e.getMessage();
                    log.error(msg);
                    applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
                }
            }
        } catch (Exception e) {
            // Do not propagate the error to the client: it would reveal whether the
            // account exists. Log and report internally instead.
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    @AuditEvent(
            attemptType = ApplicationEventType.RESET_USER_PASSWORD_ATTEMPT,
            attemptMessage = "Attempt to reset user password initiated",
            successType = ApplicationEventType.RESET_USER_PASSWORD_SUCCESS,
            successMessage = "User password reset successfully"
    )
    @PostMapping(path = "/account/reset-password/finish")
    public void finishPasswordReset(@RequestBody KeyAndPasswordVM keyAndPassword) {

        validatePasswordLength(keyAndPassword.getNewPassword());
        userService.completePasswordReset(keyAndPassword.getNewPassword(), keyAndPassword.getKey());

    }

    private void validatePasswordLength(String password) {
        if (!StringUtils.hasText(password) || password.length() < ManagedUserVM.PASSWORD_MIN_LENGTH ||
                password.length() > ManagedUserVM.PASSWORD_MAX_LENGTH)
            throw new InvalidPasswordException();
    }
}
