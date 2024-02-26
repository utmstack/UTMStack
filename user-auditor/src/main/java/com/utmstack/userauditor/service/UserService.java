package com.utmstack.userauditor.service;

import com.utmstack.userauditor.model.*;
import com.utmstack.userauditor.model.winevent.EventLog;
import com.utmstack.userauditor.repository.UserAttributeRepository;
import com.utmstack.userauditor.repository.UserRepository;
import com.utmstack.userauditor.repository.UserSourceRepository;
import com.utmstack.userauditor.service.elasticsearch.ElasticsearchService;
import com.utmstack.userauditor.service.interfaces.Source;
import com.utmstack.userauditor.service.type.EventType;
import com.utmstack.userauditor.service.type.WindowsAttributes;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;

/**
 * Service Implementation for managing {@link UserService}.
 */
@Service
@Transactional
@RequiredArgsConstructor
@Slf4j
public class UserService {

    private final UserRepository userRepository;
    private final UserSourceRepository userSourceRepository;
    private final List<Source> sources;

    /**
     * Get all the UserService.
     *
     * @param pageable the pagination information.
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public Page<User> findAll(Pageable pageable) {
        log.debug("Request to get all UserService");
        return userRepository.findAll(pageable);
    }

    /**
     * Get one UserService by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<User> findOne(Long id) {
        log.debug("Request to get UserService : {}", id);
        return userRepository.findById(id);
    }

    public Page<User> findBySource(Pageable pageable, Long sourceId) {
        log.debug("Request to get UserService : {}", sourceId);
        return userRepository.findAllBySourceId(pageable, sourceId);
    }

    @Scheduled(fixedDelayString = "${interval}")
    public void findUsers() {

        List<UserSource> logsSources = userSourceRepository.findAllByActiveIsTrue();
        logsSources.forEach(logSource -> {
            Source source = sources.stream()
                    .filter(s -> s.getType().getValue().equals(logSource.getIndexName()))
                    .findFirst()
                    .orElseThrow(() -> new RuntimeException("Class not found"));
            try {
                Map<String, List<EventLog>> users = source.findUsers(logSource);
                this.synchronizeUsers(users, logSource);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        });
    }

    private void synchronizeUsers(Map<String, List<EventLog>> users, UserSource userSource) {
        users.forEach((s, eventLogs) -> {
            User user = userRepository.findByName(s).orElse(new User());

            if (user.getId() != null) {
                if (this.isDeleteEvent(eventLogs)) {
                    user.setActive(false);
                }

                userRepository.save(user);
            } else {
                User u = User.builder()
                        .source(userSource)
                        .audit(new Audit())
                        .name(s)
                        .build();

                List<UserAttribute> userAttributes = this.getUserCreatedEvent(eventLogs) != null ? this.synchronizeAttributes(u, getUserCreatedEvent(eventLogs)): new ArrayList<>();
                u.setAttributes(userAttributes);

                userRepository.save(u);
            }
        });
    }

    private List<UserAttribute> synchronizeAttributes(User user, EventLog eventLog) {
        user.setSid(eventLog.logx.wineventlog.eventData.targetUserSid);

        List<UserAttribute> attributes = new ArrayList<>();

        for (WindowsAttributes attribute : WindowsAttributes.values()) {
            attributes.add(this.createUserAttributes(user, attribute, eventLog));
        }

        return attributes;
    }

    private UserAttribute createUserAttributes(User user, WindowsAttributes attribute, EventLog eventLog) {
        return UserAttribute
                .builder()
                .attributeKey(attribute.getValue())
                .user(user)
                .attributeValue(this.attributeByKey(attribute, eventLog))
                .audit(new Audit())
                .build();
    }

    private boolean isDeleteEvent(List<EventLog> eventLog) {
        return eventLog.stream().anyMatch(e -> e.logx.wineventlog.eventId == EventType.USER_DELETED.getEventId());
    }

    private EventLog getUserCreatedEvent(List<EventLog> eventLog) {
        return eventLog.stream().filter(s -> (s.logx.wineventlog.eventId == EventType.USER_CREATED.getEventId() || (s.logx.wineventlog.eventId == EventType.USER_LAST_LOGON.getEventId()))).findFirst().orElse(null);
    }

    private String attributeByKey(WindowsAttributes key, EventLog eventLog) {
        switch (key) {
            case SAMAccountName:
                return eventLog.logx.wineventlog.eventData.targetUserName;

            case ObjectSID:
                return eventLog.logx.wineventlog.eventData.targetUserSid;

            case CreatedAt:
                return eventLog.logx.wineventlog.event.code.equals(String.valueOf(EventType.USER_CREATED.getEventId())) ? eventLog.logx.wineventlog.event.created : "";

            case LastLogon:
                return eventLog.logx.wineventlog.event.code.equals(String.valueOf(EventType.USER_LAST_LOGON.getEventId())) ? eventLog.logx.wineventlog.event.created : "";

            default:
                return "";
        }
    }
}
