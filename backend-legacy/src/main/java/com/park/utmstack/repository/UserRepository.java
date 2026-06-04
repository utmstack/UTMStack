package com.park.utmstack.repository;

import com.park.utmstack.domain.User;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

/**
 * Spring Data JPA repository for the User entity.
 */
@Repository
public interface UserRepository extends JpaRepository<User, Long> {

    Optional<User> findOneByActivationKey(String activationKey);

    List<User> findAllByActivatedIsFalseAndCreatedDateBefore(Instant dateTime);

    Optional<User> findOneByResetKey(String resetKey);

    Optional<User> findOneByEmailIgnoreCase(String email);

    Optional<User> findOneByLogin(String login);

    @Query("select u from User u where upper(u.login) like concat('%', UPPER(:login) , '%') and (u.fsManager is null or u.fsManager is false)")
    Page<User> findAllByLoginLike(Pageable pageable, @Param("login") String login);

    @Query(nativeQuery = true, value = "SELECT count(jhi_user.id) FROM jhi_user WHERE jhi_user.id IN (SELECT jhi_user_authority.user_id FROM jhi_user_authority WHERE jhi_user_authority.authority_name = 'ROLE_ADMIN')")
    int countAdmins();

    @Query(nativeQuery = true, value = "SELECT jhi_user.* FROM jhi_user WHERE jhi_user.id IN (SELECT jhi_user_authority.user_id FROM jhi_user_authority WHERE jhi_user_authority.authority_name = 'ROLE_ADMIN')")
    List<User> findAllAdmins();

    @Query("SELECT u FROM User u JOIN FETCH u.authorities a WHERE a.name = 'ROLE_ADMIN' AND u.activated = true")
    List<User> findAdminUsers();

    @EntityGraph(attributePaths = "authorities")
    Optional<User> findOneWithAuthoritiesById(Long id);

    @EntityGraph(attributePaths = "authorities")
    Optional<User> findOneWithAuthoritiesByLogin(String login);

    @EntityGraph(attributePaths = "authorities")
    Optional<User> findOneWithAuthoritiesByEmail(String email);

    Page<User> findAllByLoginNotAndFsManagerIsNullOrFsManagerIsFalse(Pageable pageable, String login);

    List<User> findUserByIdIn(List<Long> ids);
}
