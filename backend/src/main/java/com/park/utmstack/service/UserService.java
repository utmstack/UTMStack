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
import com.park.utmstack.util.exceptions.CurrentUserLoginNotFoundException;
import com.park.utmstack.web.rest.errors.InvalidPasswordException;
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
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.Set;
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

    public Optional<User> completePasswordReset(String newPassword, String key) {
        log.debug("Reset user password for reset key {}", key);
        return userRepository.findOneByResetKey(key).filter(
            user -> user.getResetDate().isAfter(Instant.now().minusSeconds(86400))).map(user -> {
            user.setPassword(passwordEncoder.encode(newPassword));
            user.setResetKey(null);
            user.setResetDate(null);
            return user;
        });
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
            String encryptedPassword = passwordEncoder.encode(RandomUtil.generatePassword());
            user.setPassword(encryptedPassword);
            user.setResetKey(RandomUtil.generateResetKey());
            user.setResetDate(Instant.now());
            user.setActivated(false);
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
            } else {
                user.setLogin(Constants.FS_USER);
                user.setFirstName("Federation");
                user.setLastName("Service Client");
                user.setEmail(Constants.FS_USER + "@localhost");
                user.setLangKey(Constants.DEFAULT_LANGUAGE);
                user.setPassword(passwordEncoder.encode(password));
                user.setActivated(true);
                Set<Authority> authorities = Stream.of(AuthoritiesConstants.USER).map(authorityRepository::findById).filter(
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

    public User updateUserTfaSecret(String userLogin, String tfaSecret) throws Exception {
        final String ctx = CLASS_NAME + ".updateUserTfaSecret";
        try {
            User user = userRepository.findOneByLogin(userLogin)
                .orElseThrow(() -> new Exception(String.format("User %1$s not found", userLogin)));
            user.setTfaSecret(tfaSecret);
            return userRepository.save(user);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public void deleteUser(String login) {
        String ctx = CLASS_NAME + ".deleteUser";
        Optional<User> user = userRepository.findOneByLogin(login);
        if (user.isPresent()) {
            User usr = user.get();
            userRepository.delete(usr);
            log.debug("Deleted User: {}", usr);
        }
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
        return userRepository.findOneWithAuthoritiesByLogin(userLogin).orElseThrow(() -> new CurrentUserLoginNotFoundException(String.format("No user with login %1$s was found", userLogin)));
    }
}
