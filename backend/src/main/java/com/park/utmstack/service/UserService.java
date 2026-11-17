package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.Authority;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.repository.AuthorityRepository;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.AuthoritiesConstants;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UserDTO;
import com.park.utmstack.service.util.RandomUtil;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.exceptions.CurrentUserLoginNotFoundException;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.errors.InvalidPasswordException;
import com.park.utmstack.web.rest.errors.ResetKeyExpiredException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * Service class for managing users.
 */
@Service
@Transactional
public class UserService {

    private final Logger log = LoggerFactory.getLogger(UserService.class);
    private final String CLASS_NAME = "UserService";

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final AuthorityRepository authorityRepository;
    private final ApplicationEventService eventService;

    public UserService(UserRepository userRepository, PasswordEncoder passwordEncoder,
                       AuthorityRepository authorityRepository, ApplicationEventService eventService) {
        this.userRepository = userRepository;
        this.passwordEncoder = passwordEncoder;
        this.authorityRepository = authorityRepository;
        this.eventService = eventService;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASS_NAME + ".init";
        try {
            User admin = userRepository.findById(1L).orElseThrow(() ->
                new RuntimeException("Couldn't found de default admin user"));

            if (admin.getDefaultPassword())
                return;

            String dbPass = System.getenv("DB_PASS");
            if (!StringUtils.hasText(dbPass))
                throw new Exception("Environment variable DB_PASS is missing or his value is null or empty");

            admin.setDefaultPassword(true);
            admin.setPassword(passwordEncoder.encode(dbPass));
            userRepository.save(admin);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    public void completePasswordReset(String newPassword, String key) {
        final String ctx = CLASS_NAME + ".completePasswordReset";
        log.debug("{}: Processing password reset request", ctx);

        Optional<User> userOptional = userRepository.findOneByResetKey(key);

        if (userOptional.isEmpty()) {
            log.info("{}: No user found with reset key", ctx);
            throw new CurrentUserLoginNotFoundException(
                "The password reset link is invalid or no longer exists. Please request a new password reset."
            );
        }

        User user = userOptional.get();
        Instant resetDeadline = Instant.now().minusSeconds(86400);

        if (!user.getResetDate().isAfter(resetDeadline)) {
            log.error("{}: Reset key expired for user: {}", ctx, user.getLogin());
            throw new ResetKeyExpiredException(
                "The password reset link has expired. Password reset links are valid for 24 hours. Please request a new one."
            );
        }

        user.setPassword(passwordEncoder.encode(newPassword));
        user.setResetKey(null);
        user.setResetDate(null);
        user.setActivated(true);

        log.info("{}: Password reset completed successfully for user: {}", ctx, user.getLogin());
    }

    public Optional<User> requestPasswordReset(String mail) {
        return userRepository.findOneByEmailIgnoreCase(mail).filter(User::getActivated).map(user -> {
            user.setResetKey(RandomUtil.generateResetKey());
            user.setResetDate(Instant.now());
            return user;
        });
    }

    private boolean removeNonActivatedUser(User existingUser) {
        if (existingUser.getActivated()) {
            return false;
        }
        userRepository.delete(existingUser);
        userRepository.flush();
        return true;
    }

    public User createUser(UserDTO userDTO) {
        return createUser(userDTO, null);
    }

    /**
     * Create a user. When a local password is provided the user is activated
     * immediately, no email/activation step is needed.
     */
    public User createUser(UserDTO userDTO, String localPassword) {
        String ctx = CLASS_NAME + ".createUser";
        try {
            User user = new User();
            user.setLogin(userDTO.getLogin().toLowerCase());
            user.setFirstName(userDTO.getFirstName());
            user.setLastName(userDTO.getLastName());
            user.setEmail(userDTO.getEmail().toLowerCase());
            user.setImageUrl(userDTO.getImageUrl());
            if (userDTO.getLangKey() == null) {
                user.setLangKey(Constants.DEFAULT_LANGUAGE); // default language
            } else {
                user.setLangKey(userDTO.getLangKey());
            }
            if (StringUtils.hasText(localPassword)) {
                user.setPassword(passwordEncoder.encode(localPassword));
                user.setActivated(true);
            } else {
                String encryptedPassword = passwordEncoder.encode(RandomUtil.generatePassword());
                user.setPassword(encryptedPassword);
                user.setResetKey(RandomUtil.generateResetKey());
                user.setResetDate(Instant.now());
                user.setActivated(false);
            }
            if (userDTO.getAuthorities() != null) {
                Set<Authority> authorities = userDTO.getAuthorities().stream().map(authorityRepository::findById).filter(
                    Optional::isPresent).map(Optional::get).collect(Collectors.toSet());
                user.setAuthorities(authorities);
            }
            userRepository.save(user);
            log.debug("Created Information for User: {}", user);
            return user;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public void createFederationServiceUser(String password) {
        String ctx = CLASS_NAME + ".createFederationServiceUser";
        try {
            User user = userRepository.findOneByLogin(Constants.FS_USER).orElse(new User());

            if (!Objects.isNull(user.getId())) {
                user.setPassword(passwordEncoder.encode(password));
                if (user.getAuthorities().stream().noneMatch(a -> a.getName().equals(AuthoritiesConstants.ADMIN)))
                    user.getAuthorities().add(authorityRepository.findById(AuthoritiesConstants.ADMIN).orElseThrow());
            } else {
                user.setLogin(Constants.FS_USER);
                user.setFirstName("Federation");
                user.setLastName("Service Client");
                user.setEmail(Constants.FS_USER + "@localhost");
                user.setLangKey(Constants.DEFAULT_LANGUAGE);
                user.setPassword(passwordEncoder.encode(password));
                user.setActivated(true);
                Set<Authority> authorities = Stream.of(AuthoritiesConstants.USER, AuthoritiesConstants.ADMIN).map(authorityRepository::findById).filter(
                    Optional::isPresent).map(Optional::get).collect(Collectors.toSet());
                user.setAuthorities(authorities);
                user.setFsManager(true);
            }

            userRepository.save(user);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Update basic information (first name, last name, email, language) for the current user.
     *
     * @param firstName first name of user
     * @param lastName  last name of user
     * @param email     email id of user
     * @param langKey   language key
     * @param imageUrl  image URL of user
     */
    public void updateUser(String firstName, String lastName, String email, String langKey, String imageUrl) {
        SecurityUtils.getCurrentUserLogin().flatMap(userRepository::findOneByLogin).ifPresent(user -> {
            user.setFirstName(firstName);
            user.setLastName(lastName);
            user.setEmail(email.toLowerCase());
            user.setLangKey(langKey);
            user.setImageUrl(imageUrl);
            log.debug("Changed Information for User: {}", user);
        });
    }

    /**
     * Update all information for a specific user, and return the modified user.
     *
     * @param userDTO user to update
     * @return updated user
     */
    public Optional<UserDTO> updateUser(UserDTO userDTO) {
        String ctx = CLASS_NAME + ".updateUser";
        Optional<User> userOptional = userRepository.findById(userDTO.getId());

        return Optional.of(userOptional).filter(Optional::isPresent).map(Optional::get).map(user -> {
            user.setLogin(userDTO.getLogin().toLowerCase());
            user.setFirstName(userDTO.getFirstName());
            user.setLastName(userDTO.getLastName());
            user.setEmail(userDTO.getEmail().toLowerCase());
            user.setImageUrl(userDTO.getImageUrl());
            user.setActivated(userDTO.isActivated());
            user.setLangKey(userDTO.getLangKey());
            Set<Authority> managedAuthorities = user.getAuthorities();
            managedAuthorities.clear();
            userDTO.getAuthorities().stream().map(authorityRepository::findById).filter(Optional::isPresent).map(
                Optional::get).forEach(managedAuthorities::add);
            log.debug("Changed Information for User: {}", user);
            return user;
        }).map(UserDTO::new);
    }

    public void updateUserTfaSecret(String userLogin, String tfaSecret, String tfaMethod) {
            User user = userRepository.findOneByLogin(userLogin)
                .orElseThrow(() -> new NoSuchElementException(String.format("User %1$s not found", userLogin)));
            user.setTfaMethod(tfaMethod);
            user.setTfaSecret(encryptTfaSecret(tfaSecret, encryptionKey()));

            userRepository.save(user);
    }

    /**
     * Encrypts a TFA secret at rest so it is never stored (or backed up) as the
     * raw seed. TOTP/Email seeds are base32 and the house cipher output is
     * base64, so the {@code enc::} prefix makes "encrypted" unambiguous and lets
     * readers distinguish legacy plaintext rows. Re-encrypting an already
     * encrypted value is a no-op (the migration runner relies on this).
     */
    public static String encryptTfaSecret(String secret, String key) {
        if (secret == null || secret.isBlank() || isEncryptedTfaSecret(secret)) {
            return secret;
        }
        return Constants.TFA_SECRET_ENCRYPTED_PREFIX + CipherUtil.encrypt(secret, key);
    }

    /**
     * Decrypts a stored TFA secret for use in code generation. Values without
     * the {@code enc::} prefix are legacy plaintext seeds and are returned
     * as-is, so TFA keeps working even before the one-time migration has run.
     */
    public static String decryptTfaSecret(String stored, String key) {
        if (stored == null || !isEncryptedTfaSecret(stored)) {
            return stored;
        }
        return CipherUtil.decrypt(stored.substring(Constants.TFA_SECRET_ENCRYPTED_PREFIX.length()), key);
    }

    /**
     * Instance convenience for the TFA code-generation read path: resolves the
     * encryption key (failing fast if it is not configured) and decrypts.
     */
    public String resolveTfaSecret(String stored) {
        return decryptTfaSecret(stored, encryptionKey());
    }

    public static boolean isEncryptedTfaSecret(String stored) {
        return stored != null && stored.startsWith(Constants.TFA_SECRET_ENCRYPTED_PREFIX);
    }

    private String encryptionKey() {
        String key = System.getenv(Constants.ENV_ENCRYPTION_KEY);
        if (!StringUtils.hasText(key)) {
            throw new IllegalStateException(CLASS_NAME + ": " + Constants.ENV_ENCRYPTION_KEY + " is not configured");
        }
        return key;
    }

    /**
     * One-time migration: re-encrypts legacy plaintext TFA secrets at rest.
     * Rows already carrying the {@code enc::} prefix are skipped, so the runner
     * is idempotent and safe to execute on every boot. Per-row failures are
     * logged and skipped so a single bad row cannot block startup; the lazy
     * plaintext fallback in {@link #decryptTfaSecret(String, String)} keeps
     * those users' TFA functional until the row migrates.
     *
     * @return the number of rows migrated
     */
    public int migrateLegacyTfaSecrets() {
        String key = encryptionKey();
        int migrated = 0;
        for (User user : userRepository.findAll()) {
            String stored = user.getTfaSecret();
            if (stored == null || stored.isBlank() || isEncryptedTfaSecret(stored)) {
                continue;
            }
            try {
                user.setTfaSecret(encryptTfaSecret(stored, key));
                userRepository.save(user);
                migrated++;
            } catch (Exception e) {
                log.error(CLASS_NAME + ".migrateLegacyTfaSecrets: failed to encrypt secret for user {}: {}",
                        user.getLogin(), e.getMessage());
            }
        }
        return migrated;
    }

    public void deleteUser(String login) {
        String ctx = CLASS_NAME + ".deleteUser";
        User user = userRepository.findOneByLogin(login)
                .orElseThrow(() -> new NoSuchElementException(String.format("User %1$s not found", login)));

        if (user.getAuthorities().stream().anyMatch(authority -> authority.getName().equals("ROLE_ADMIN")) && userRepository.countAdmins() == 1) {
            throw new BadRequestAlertException(ctx, "Cannot delete the last admin user.", UserService.class.toString());
        }

        userRepository.delete(user);
        log.debug("Deleted User: {}", user);
    }

    public void changePassword(String currentClearTextPassword, String newPassword) {
        SecurityUtils.getCurrentUserLogin().flatMap(userRepository::findOneByLogin).ifPresent(user -> {
            String currentEncryptedPassword = user.getPassword();
            if (!passwordEncoder.matches(currentClearTextPassword, currentEncryptedPassword)) {
                throw new InvalidPasswordException();
            }
            String encryptedPassword = passwordEncoder.encode(newPassword);
            user.setPassword(encryptedPassword);
            log.debug("Changed password for User: {}", user);
        });
    }

    @Transactional(readOnly = true)
    public Page<UserDTO> getAllManagedUsers(Pageable pageable, String login) {
        if (StringUtils.hasText(login))
            return userRepository.findAllByLoginLike(pageable, login).map(UserDTO::new);
        return userRepository.findAllByLoginNotAndFsManagerIsNullOrFsManagerIsFalse(pageable, Constants.ANONYMOUS_USER).map(UserDTO::new);
    }

    @Transactional(readOnly = true)
    public List<UserDTO> getAllUsersIn(List<Long> ids) {
        return userRepository.findUserByIdIn(ids).stream().map(UserDTO::new).collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public Optional<User> getUserWithAuthoritiesByLogin(String login) {
        return userRepository.findOneWithAuthoritiesByLogin(login);
    }

    @Transactional(readOnly = true)
    public Optional<User> getUserWithAuthorities(Long id) {
        return userRepository.findOneWithAuthoritiesById(id);
    }

    @Transactional(readOnly = true)
    public Optional<User> getUserWithAuthorities() {
        return SecurityUtils.getCurrentUserLogin().flatMap(userRepository::findOneWithAuthoritiesByLogin);
    }

    public Page<UserDTO> getUsersByLogin(Pageable pageable, String login) {
        return userRepository.findAllByLoginLike(pageable, login).map(UserDTO::new);
    }

    /**
     * Not activated users should be automatically deleted after 3 days.
     * <p>
     * This is scheduled to get fired everyday, at 01:00 (am).
     */
    @Scheduled(cron = "0 0 1 * * ?")
    public void removeNotActivatedUsers() {
        userRepository.findAllByActivatedIsFalseAndCreatedDateBefore(Instant.now().minus(3, ChronoUnit.DAYS)).forEach(
            user -> {
                log.debug("Deleting not activated user {}", user.getLogin());
                userRepository.delete(user);
            });
    }

    /**
     * @return a list of all the authorities
     */
    public List<String> getAuthorities() {
        return authorityRepository.findAll().stream().map(Authority::getName).collect(Collectors.toList());
    }

    public User getCurrentUserLogin() {
        String userLogin = SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new CurrentUserLoginNotFoundException("No current user login was found"));
        return userRepository.findOneWithAuthoritiesByLogin(userLogin)
                .orElseThrow(() -> new CurrentUserLoginNotFoundException(String.format("No user with login %1$s was found", userLogin)));
    }
}
