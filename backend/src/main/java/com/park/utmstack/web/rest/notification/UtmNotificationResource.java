package com.park.utmstack.web.rest.notification;

import com.park.utmstack.domain.notification.NotificationFilters;
import com.park.utmstack.domain.notification.NotificationStatus;
import com.park.utmstack.domain.notification.UtmNotification;
import com.park.utmstack.service.dto.notification.NotificationDTO;
import com.park.utmstack.service.dto.notification.UtmNotificationMapper;
import com.park.utmstack.service.notification.UtmNotificationService;
import lombok.extern.slf4j.Slf4j;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;


@RestController
@Slf4j
@RequestMapping("/api/notifications")
public class UtmNotificationResource {

    private static final String CLASSNAME = "UtmNotificationResource";

    private final UtmNotificationService notificationService;

    private final UtmNotificationMapper notificationMapper;


    public UtmNotificationResource(UtmNotificationService notificationService, UtmNotificationMapper notificationMapper) {
        this.notificationService = notificationService;
        this.notificationMapper = notificationMapper;
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
     * Retrieves a paginated list of notifications based on the provided filters.
     *
     * @param filters  the filters to apply for the notification search
     * @param pageable the pagination information (page number, size, sort order)
     * @return a ResponseEntity containing a Page of NotificationDTOs
     */
    @GetMapping
    public ResponseEntity<Page<NotificationDTO>> getAllNotifications(
            @ParameterObject NotificationFilters filters,
            @ParameterObject Pageable pageable) {
        return ResponseEntity.ok(notificationService.getNotifications(filters, pageable));
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
    @PutMapping("/{id}/read")
    public ResponseEntity<?> updateNotificationReadStatus(@PathVariable Long id, @RequestParam boolean read) {
        final String ctx = CLASSNAME + ".updateNotificationReadStatus";
        try{
            UtmNotification notification = notificationService.updateNotificationReadStatus(id, read);
            return ResponseEntity.ok(notificationMapper.toDto(notification));
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage());
            throw e;
        }
    }

    /**
     * Update the "status" status of a notification.
     *
     * @param id the ID of the notification to update
     * @param status the new status
     * @return the updated notification
     */
    @PutMapping("/{id}/status")
    public ResponseEntity<?> updateNotificationStatus(@PathVariable Long id, @RequestParam NotificationStatus status) {
        final String ctx = CLASSNAME + ".updateNotificationStatus";
        try{
            UtmNotification notification = notificationService.updateNotificationStatus(id, status);
            return ResponseEntity.ok(notificationMapper.toDto(notification));
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage());
            throw e;
        }
    }

    /**
     * Get the count of unread notifications.
     *
     * @return the count of unread notifications
     */
    @GetMapping("/unread-count")
    public ResponseEntity<Integer> getUnreadNotificationCount() {
        return ResponseEntity.ok(notificationService.getUnreadNotifications());
    }

    /**
     * Marks all notifications as read.
     *
     * This method does not return any data.
     */
    @PutMapping("/read-all")
    public ResponseEntity<Void> markAllNotificationsAsRead() {
        notificationService.markAllNotificationsAsRead();
        return ResponseEntity.ok().build();
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
        } catch (Exception e) {
            log.error(ctx + ": " + e.getMessage());
            throw e;
        }
    }
}
