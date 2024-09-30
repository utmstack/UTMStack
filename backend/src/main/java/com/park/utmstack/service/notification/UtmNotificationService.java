package com.park.utmstack.service.notification;

import com.park.utmstack.domain.notification.*;
import com.park.utmstack.repository.notification.UtmNotificationRepository;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.dto.notification.NotificationDTO;
import com.park.utmstack.service.dto.notification.UtmNotificationMapper;
import com.park.utmstack.service.mail_config.MailConfigService;
import com.park.utmstack.util.AlertUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import javax.mail.MessagingException;
import java.time.LocalDateTime;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

@RequiredArgsConstructor
@Service
@Transactional
public class UtmNotificationService {

    private final UtmNotificationRepository notificationRepository;

    private final UtmNotificationMapper notificationMapper;

    private final AlertUtil alertUtil;

    private final MailService mailService;

    private final MailConfigService mailConfigService;

    public UtmNotification saveNotification(UtmNotification notification) {
        notification.setCreatedAt(LocalDateTime.now());
        notification.setRead(false);
        notification.setStatus(NotificationStatus.ACTIVE);
        return notificationRepository.save(notification);
    }

    public Page<NotificationDTO> getNotifications(NotificationFilters filters, Pageable pageable) {

        Page<UtmNotification> page = notificationRepository.searchByFilters(filters.getSource(),
                        filters.getType(),
                        filters.getStatus(),
                        filters.getFrom(),
                        filters.getTo(),
                        pageable);

        List<NotificationDTO> notificationDTOS = page.getContent()
                .stream()
                .map(notificationMapper::toDto)
                .collect(Collectors.toList());


        return new PageImpl<>(notificationDTOS, pageable, page.getTotalElements());
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

    public UtmNotification updateNotificationStatus(Long id, NotificationStatus status) {
        UtmNotification notification = getNotificationById(id);
        notification.setStatus(status);
        notification.setRead(true);
        return notificationRepository.save(notification);
    }

    public void markAllNotificationsAsRead() {
        this.notificationRepository.updateUnreadNotifications(true);
    }


    public void deleteNotification(Long id) {
        notificationRepository.deleteById(id);
    }

    public int getUnreadNotifications() {
        return this.notificationRepository.countUtmNotificationByReadIsFalseAndStatusIsLike(NotificationStatus.ACTIVE);
    }

   /* @Scheduled(fixedDelay = 1800000)
    public void loadOpenAlerts() {
        if (this.countAlert == -1) {
            this.countAlert = this.alertUtil.countAlertsByStatus(2);
            if(this.countAlert > 0){
                this.sendNotification(String.format("There are %1$s pending alerts to manage", this.countAlert), NotificationSource.ALERTS, NotificationType.INFO);
            }
        } else {
            Long latestCountAlert = this.alertUtil.countAlertsByStatus(2, LocalDateTime.now().minusMinutes(30), LocalDateTime.now());
            Long totalAlert = this.alertUtil.countAlertsByStatus(2);

            if (this.countAlert != totalAlert && latestCountAlert > 0) {
                this.countAlert = totalAlert;
                this.sendNotification(String.format("There are %1$s pending alerts to manage", this.countAlert), NotificationSource.ALERTS, NotificationType.INFO);
            }
        }
    }*/

    @Scheduled(fixedDelay = 4320000)
    public void checkEmailConfig() {
        try {
            this.mailService.getJavaMailSender();
        } catch (MessagingException e) {
            this.sendNotification(e.getMessage(), NotificationSource.EMAIL_SETTING, NotificationType.ERROR);
        }
    }

    public void sendNotification(String message, NotificationSource source, NotificationType type){
        saveNotification(UtmNotification.builder()
                .message(message)
                .source(source)
                .createdAt(LocalDateTime.now())
                .read(false)
                .type(type)
                .build());
    }
}

