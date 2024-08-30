package com.park.utmstack.web.rest.notification;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.notification.UtmNotification;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.collectors.dto.ErrorResponse;
import com.park.utmstack.service.dto.notification.NotificationDTO;
import com.park.utmstack.service.dto.notification.UtmNotificationMapper;
import com.park.utmstack.service.notification.UtmNotificationService;
import com.park.utmstack.util.UtilResponse;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.persistence.EntityNotFoundException;
import javax.validation.Valid;
import java.util.List;
import java.util.stream.Collectors;

@RestController
@Slf4j
@RequestMapping("/api/notifications")
public class UtmNotificationResource {

    private static final String CLASSNAME = "UtmNotificationResource";

    private final UtmNotificationService notificationService;

    private final UtmNotificationMapper notificationMapper;

    private final ApplicationEventService applicationEventService;

    public UtmNotificationResource(UtmNotificationService notificationService, UtmNotificationMapper notificationMapper, ApplicationEventService applicationEventService) {
        this.notificationService = notificationService;
        this.notificationMapper = notificationMapper;
        this.applicationEventService = applicationEventService;
    }

    /**
     * Create a new notification.
     *
     * @param notificationDTO the notification to create
     * @return the created notification
     */
    @PostMapping
    public ResponseEntity<NotificationDTO> createNotification(@RequestBody @Valid NotificationDTO notificationDTO) {
        NotificationDTO createdNotification = notificationMapper.toDto(
                notificationService.saveNotification(notificationMapper.toEntity(notificationDTO)));
        return ResponseEntity.status(HttpStatus.CREATED).body(createdNotification);
    }

    /**
     * Get all notifications.
     *
     * @return the list of notifications
     */
    @GetMapping
    public ResponseEntity<Page<NotificationDTO>> getAllNotifications(Pageable pageable) {
        return ResponseEntity.ok(notificationService.getNotifications(pageable));
    }

    /**
     * Get a specific notification by ID.
     *
     * @param id the ID of the notification to retrieve
     * @return the notification
     */
    @GetMapping("/{id}")
    public ResponseEntity<?> getNotificationById(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".findByIdNotification";
        try{
            UtmNotification notification = notificationService.getNotificationById(id);
            return ResponseEntity.ok(notificationMapper.toDto(notification));
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage());
            throw e;
        }
    }

    /**
     * Update the "read" status of a notification.
     *
     * @param id the ID of the notification to update
     * @param read the new read status
     * @return the updated notification
     */
    @PatchMapping("/{id}/read")
    public ResponseEntity<?> updateNotificationReadStatus(@PathVariable Long id, @RequestParam boolean read) {
        final String ctx = CLASSNAME + ".updateNotificationReadStatus";
        try{
            UtmNotification notification = notificationService.updateNotificationReadStatus(id, read);
            return ResponseEntity.ok(notificationMapper.toDto(notification));
        } catch (EntityNotFoundException e) {
            return logAndResponse(new ErrorResponse(ctx + ": " + e.getMessage(), HttpStatus.NOT_FOUND));
        } catch (Exception e) {
            return logAndResponse(new ErrorResponse(ctx + ": " + e.getMessage(), HttpStatus.INTERNAL_SERVER_ERROR));
        }
    }

    /**
     * Delete a notification by ID.
     *
     * @param id the ID of the notification to delete
     * @return the response status
     */
    @DeleteMapping("/{id}")
    public ResponseEntity<?> deleteNotification(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteNotification";
        try {
            notificationService.deleteNotification(id);
            return ResponseEntity.ok().build();
        } catch (EntityNotFoundException e) {
            return logAndResponse(new ErrorResponse(ctx + ": " + e.getMessage(), HttpStatus.NOT_FOUND));
        } catch (Exception e) {
            return logAndResponse(new ErrorResponse(ctx + ": " + e.getMessage(), HttpStatus.INTERNAL_SERVER_ERROR));
        }
    }

    /**
     * Get notifications by user ID.
     *
     * @param userId the ID of the user
     * @return the list of notifications for the user
     */
    /*@GetMapping("/user/{userId}")
    public ResponseEntity<List<NotificationDTO>> getNotificationsByUserId(@PathVariable Long userId) {
        List<NotificationDTO> notifications = notificationService.getNotificationsByUserId(userId);
        return ResponseEntity.ok(notifications);
    }*/

    private ResponseEntity<Void> logAndResponse(ErrorResponse error) {
        log.error(error.getMessage());
        applicationEventService.createEvent(error.getMessage(), ApplicationEventType.ERROR);
        return UtilResponse.buildErrorResponse(error.getStatus(), error.getMessage());
    }
}
