package com.park.utmstack.service.dto.notification;

import com.park.utmstack.domain.notification.UtmNotification;
import org.mapstruct.Mapper;
import org.mapstruct.Mapping;


@Mapper(componentModel = "spring")
public interface UtmNotificationMapper {

    @Mapping(source = "id", target = "id")
    @Mapping(source = "source", target = "source")
    @Mapping(source = "type", target = "type")
    @Mapping(source = "message", target = "message")
    @Mapping(source = "createdAt", target = "createdAt")
    @Mapping(source = "updatedAt", target = "updatedAt")
    @Mapping(source = "read", target = "read")
    public NotificationDTO toDto(UtmNotification entity);
    public UtmNotification toEntity(NotificationDTO dto);

}
