package com.park.utmstack.repository.notification;


import com.park.utmstack.domain.notification.UtmNotification;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface UtmNotificationRepository extends JpaRepository<UtmNotification, Long> {
    int countUtmNotificationByReadIsFalse();
}
