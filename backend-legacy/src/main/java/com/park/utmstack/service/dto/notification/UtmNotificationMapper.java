package com.park.utmstack.service.dto.notification;

import com.park.utmstack.domain.notification.UtmNotification;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;


@Mapper(componentModel = "spring")
public interface UtmNotificationMapper {
    public NotificationDTO toDto(UtmNotification entity);
    public UtmNotification toEntity(NotificationDTO dto);

}
