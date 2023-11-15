package com.utmstack.userauditor.dto;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import org.springframework.data.domain.Pageable;
@Getter
@Setter
@AllArgsConstructor
@NoArgsConstructor
public class MicroserviceRequest {
    Long id;
    String criteria;
    int top;
    String index;
}
