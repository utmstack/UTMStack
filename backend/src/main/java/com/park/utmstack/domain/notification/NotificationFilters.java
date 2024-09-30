package com.park.utmstack.domain.notification;

import lombok.*;

import java.time.Instant;
import java.time.LocalDateTime;

@Data
@RequiredArgsConstructor
@AllArgsConstructor
@Getter
@Setter
public class NotificationFilters {
    NotificationSource source;
    NotificationType type;
    NotificationStatus status;
    LocalDateTime from;
    LocalDateTime to;
}
