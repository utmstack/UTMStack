package com.park.utmstack.repository.notification;


import com.park.utmstack.domain.notification.NotificationSource;
import com.park.utmstack.domain.notification.NotificationType;
import com.park.utmstack.domain.notification.UtmNotification;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;

@Repository
public interface UtmNotificationRepository extends JpaRepository<UtmNotification, Long> {
    int countUtmNotificationByReadIsFalse();

    @Query(value = "SELECT DISTINCT n FROM UtmNotification n " +
                   "WHERE " +
                    "(:source IS NULL OR (n.source LIKE :source OR lower(n.source) LIKE lower(:source))) " +
                   "AND ((:type) IS NULL OR n.type IN (:type)) " +
                   "AND ((cast(:from as timestamp) is null) and (cast(:to as timestamp) is null) or (n.createdAt BETWEEN :from AND :to)) ")
    Page<UtmNotification> searchByFilters(@Param("source") NotificationSource source,
                                          @Param("type") NotificationType type,
                                          @Param("from") LocalDateTime from,
                                          @Param("to") LocalDateTime to,
                                          Pageable pageable);
}
