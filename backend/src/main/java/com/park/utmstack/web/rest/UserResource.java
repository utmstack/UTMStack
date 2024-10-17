package com.park.utmstack.web.rest;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.AuthoritiesConstants;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UserDTO;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.errors.EmailAlreadyUsedException;
import com.park.utmstack.web.rest.errors.LoginAlreadyUsedException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * REST controller for managing users.
 * <p>
 * This class accesses the User entity, and needs to fetch its collection of authorities.
 * <p>
 * For a normal use-case, it would be better to have an eager relationship between User and Authority,
 * and send everything to the client side: there would be no View Model and DTO, a lot less code, and an outer-join
 * which would be good for performance.
 * <p>
 * We use a View Model and a DTO for 3 reasons:
 * <ul>
 * <li>We want to keep a lazy association between the user and the authorities, because people will
 * quite often do relationships with the user, and we don't want them to get the authorities all
 * the time for nothing (for performance reasons). This is the #1 goal: we should not impact our users'
 * application because of this use-case.</li>
 * <li> Not having an outer join causes n+1 requests to the database. This is not a real issue as
 * we have by default a second-level cache. This means on the first HTTP call we do the n+1 requests,
 * but then all authorities come from the cache, so in fact it's much better than doing an outer join
 * (which will get lots of data from the database, for each HTTP call).</li>
 * <li> As this manages users, for security reasons, we'd rather have a DTO layer.</li>
 * </ul>
 * <p>
 * Another option would be to have a specific JPA entity graph to handle this case.
 */
@RestController
@RequestMapping("/api")
public class UserResource {

    private static final String CLASSNAME = "UserResource";
    private final Logger log = LoggerFactory.getLogger(UserResource.class);

    private final UserService userService;
    private final UserRepository userRepository;
    private final MailService mailService;
    private final ApplicationEventService applicationEventService;

    public UserResource(UserService userService,
                        UserRepository userRepository,
                        MailService mailService,
                        ApplicationEventService applicationEventService) {
        this.userService = userService;
        this.userRepository = userRepository;
        this.mailService = mailService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /users  : Creates a new user.
     * <p>
     * Creates a new user if the login and email are not already used, and sends an
     * mail with an activation link.
     * The user needs to be activated on creation.
     *
     * @param userDTO the user to create
     * @return the ResponseEntity with status 201 (Created) and with body the new user, or with status 400 (Bad Request) if the login or email is already in use
     * @throws BadRequestAlertException 400 (Bad Request) if the login or email is already in use
     */
    @PostMapping("/users")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public ResponseEntity<User> createUser(@Valid @RequestBody UserDTO userDTO) {
        final String ctx = CLASSNAME + ".createUser";
        try {
            if (userDTO.getId() != null) {
                throw new BadRequestAlertException("A new user cannot already have an ID", "userManagement", "idexists");
                // Lowercase the user login before comparing with database
            } else if (userRepository.findOneByLogin(userDTO.getLogin()
                .toLowerCase())
                .isPresent()) {
                throw new LoginAlreadyUsedException();
            } else if (userRepository.findOneByEmailIgnoreCase(userDTO.getEmail())
                .isPresent()) {
                throw new EmailAlreadyUsedException();
            } else {
                User newUser = userService.createUser(userDTO);
                mailService.sendCreationEmail(newUser);
                return ResponseEntity.created(new URI("/api/users/" + newUser.getLogin()))
                    .headers(HeaderUtil.createAlert("A user is created with identifier " + newUser.getLogin(),
                        newUser.getLogin()))
                    .body(newUser);
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT /users : Updates an existing User.
     *
     * @param userDTO the user to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated user
     * @throws EmailAlreadyUsedException 400 (Bad Request) if the email is already in use
     * @throws LoginAlreadyUsedException 400 (Bad Request) if the login is already in use
     */
    @PutMapping("/users")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public ResponseEntity<UserDTO> updateUser(@Valid @RequestBody UserDTO userDTO) {
        final String ctx = CLASSNAME + ".updateUser";
        try {
            Optional<User> existingUser = userRepository.findOneByEmailIgnoreCase(userDTO.getEmail());
            if (existingUser.isPresent() && (!existingUser.get()
                    .getId()
                    .equals(userDTO.getId()))) {
                throw new EmailAlreadyUsedException();
            }
            existingUser = userRepository.findOneByLogin(userDTO.getLogin()
                    .toLowerCase());
            if (existingUser.isPresent() && (!existingUser.get()
                    .getId()
                    .equals(userDTO.getId()))) {
                throw new LoginAlreadyUsedException();
            }

            User user = userRepository.findOneWithAuthoritiesById(userDTO.getId())
                    .orElseThrow(() -> new NoSuchElementException(String.format("User %1$s not found", userDTO.getId().toString())));
            if (!userDTO.getAuthorities().contains("ROLE_ADMIN") &&
                    user.getAuthorities().stream().anyMatch(authority -> authority.getName().equals("ROLE_ADMIN")) && userRepository.countAdmins() == 1) {
                throw new BadRequestAlertException(ctx, "Cannot update roles for the last remaining admin user.", UserService.class.toString());
            }

            Optional<UserDTO> updatedUser = userService.updateUser(userDTO);

            return ResponseUtil.wrapOrNotFound(updatedUser,
                    HeaderUtil.createAlert("A user is updated with identifier " + userDTO.getLogin(),
                            userDTO.getLogin()));

        } catch (NoSuchElementException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.NOT_FOUND).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET /users : get all users.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and with body all users
     */
    @GetMapping("/users")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public ResponseEntity<List<UserDTO>> getAllUsers(Pageable pageable, @RequestParam(required = false) String login) {
        final String ctx = CLASSNAME + ".getAllUsers";
        try {
            final Page<UserDTO> page = userService.getAllManagedUsers(pageable, login);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/users");
            return new ResponseEntity<>(page.getContent(), headers, HttpStatus.OK);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * @return a string list of the all of the roles
     */
    @GetMapping("/users/authorities")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public List<String> getAuthorities() {
        final String ctx = CLASSNAME + ".getAuthorities";
        try {
            return userService.getAuthorities();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * GET /users/:login : get the "login" user.
     *
     * @param login the login of the user to find
     * @return the ResponseEntity with status 200 (OK) and with body the "login" user, or with status 404 (Not Found)
     */
    @GetMapping("/users/{login:" + Constants.LOGIN_REGEX + "}")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public ResponseEntity<UserDTO> getUser(@PathVariable String login) {
        final String ctx = CLASSNAME + ".getUser";
        try {
            return ResponseUtil.wrapOrNotFound(userService.getUserWithAuthoritiesByLogin(login)
                .map(UserDTO::new));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE /users/:login : delete the "login" User.
     *
     * @param login the login of the user to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/users/{login:" + Constants.LOGIN_REGEX + "}")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public ResponseEntity<Void> deleteUser(@PathVariable String login) {
        final String ctx = CLASSNAME + ".deleteUser";
        try {
            userService.deleteUser(login);
            return ResponseEntity.ok()
                    .headers(HeaderUtil.createAlert("A user is deleted with identifier " + login, login))
                    .build();
        } catch (NoSuchElementException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.NOT_FOUND).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.BAD_REQUEST).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/users/filter/{login}")
    @PreAuthorize("hasRole(\"" + AuthoritiesConstants.ADMIN + "\")")
    public ResponseEntity<List<UserDTO>> getUsersByLogin(@PathVariable String login, Pageable pageable) {
        final String ctx = CLASSNAME + ".getUsersByLogin";
        try {
            Page<UserDTO> usersOpt = userService.getUsersByLogin(pageable, login);
            List<UserDTO> result = usersOpt.getContent();

            if (!CollectionUtils.isEmpty(result)) {
                Optional<String> userLogin = SecurityUtils.getCurrentUserLogin();
                result = result.stream()
                    .filter(usr -> !usr.getLogin()
                        .equals(userLogin.orElse("")))
                    .collect(Collectors.toList());
            }
            return ResponseEntity.ok()
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}
