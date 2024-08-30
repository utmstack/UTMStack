package com.park.utmstack.service.notification;

import com.park.utmstack.domain.notification.UtmNotification;
import com.park.utmstack.repository.notification.UtmNotificationRepository;
import com.park.utmstack.service.dto.notification.NotificationDTO;
import com.park.utmstack.service.dto.notification.UtmNotificationMapper;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;

import javax.persistence.EntityNotFoundException;
import java.time.Instant;
import java.time.LocalDateTime;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

@Service
public class UtmNotificationService {

    private final UtmNotificationRepository notificationRepository;

    private final UtmNotificationMapper notificationMapper;

    public UtmNotificationService(UtmNotificationRepository notificationRepository,
                                  UtmNotificationMapper notificationMapper) {
        this.notificationRepository = notificationRepository;
        this.notificationMapper = notificationMapper;
    }

    public UtmNotification saveNotification(UtmNotification notification) {
        notification.setCreatedAt(LocalDateTime.now());
        notification.setRead(false);
        return notificationRepository.save(notification);
    }

    public Page<NotificationDTO> getNotifications(Pageable pageable) {
        Page<UtmNotification> page = notificationRepository.findAll(pageable);
        List<NotificationDTO> notificationDTOS = notificationRepository.findAll(pageable).getContent()
                .stream().map(notificationMapper::toDto)
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
}

