package com.park.utmstack.service.notification;

import com.park.utmstack.domain.notification.NotificationFilters;
import com.park.utmstack.domain.notification.NotificationSource;
import com.park.utmstack.domain.notification.NotificationType;
import com.park.utmstack.domain.notification.UtmNotification;
import com.park.utmstack.repository.notification.UtmNotificationRepository;
import com.park.utmstack.service.dto.notification.NotificationDTO;
import com.park.utmstack.service.dto.notification.UtmNotificationMapper;
import com.park.utmstack.util.AlertUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

@RequiredArgsConstructor
@Service
public class UtmNotificationService {

    private final UtmNotificationRepository notificationRepository;

    private final UtmNotificationMapper notificationMapper;

    private final AlertUtil alertUtil;

    private long countAlert = -1;

    public UtmNotification saveNotification(UtmNotification notification) {
        notification.setCreatedAt(LocalDateTime.now());
        notification.setRead(false);
        return notificationRepository.save(notification);
    }

    public Page<NotificationDTO> getNotifications(NotificationFilters filters, Pageable pageable) {
        List<NotificationDTO> notificationDTOS = notificationRepository.searchByFilters(filters.getSource(),
                        filters.getType(),
                        filters.getFrom(),
                        filters.getTo())
                .stream()
                .map(notificationMapper::toDto)
                .collect(Collectors.toList());
        return new PageImpl<>(notificationDTOS, pageable, notificationDTOS.size());
    }

    public UtmNotification getNotificationById(Long id) {
        return notificationRepository.findById(id)
                .orElseThrow(() -> new NoSuchElementException("Notification not found with id: " + id));
    }

    public UtmNotification updateNotificationReadStatus(Long id, boolean read) {
        UtmNotification notification = getNotificationById(id);
        notification.setRead(read);
        return notificationRepository.save(notification);
    }

    public void deleteNotification(Long id) {
        notificationRepository.deleteById(id);
    }

    public int getUnreadNotifications() {
        return this.notificationRepository.countUtmNotificationByReadIsFalse();
    }

    @Scheduled(fixedDelay = 1800000)
    public void loadOpenAlerts() {
        if (this.countAlert == -1) {
            this.countAlert = this.alertUtil.countAlertsByStatus(2);
            if(this.countAlert > 0){
                this.sendNotification();
            }
        } else {
            Long latestCountAlert = this.alertUtil.countAlertsByStatus(2, LocalDateTime.now().minusMinutes(30), LocalDateTime.now());
            Long totalAlert = this.alertUtil.countAlertsByStatus(2);

            if (this.countAlert != totalAlert && latestCountAlert > 0) {
                this.countAlert = totalAlert;
                this.sendNotification();
            }
        }
    }

    private void sendNotification(){
        saveNotification(UtmNotification.builder()
                .message(String.format("There are %1$s pending alerts to manage", this.countAlert))
                .source(NotificationSource.ALERTS)
                .createdAt(LocalDateTime.now())
                .read(false)
                .type(NotificationType.INFO)
                .build());
    }
}

