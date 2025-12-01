package com.park.utmstack.service.dto.threat_management;

import com.park.utmstack.domain.shared_types.alert.Side;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class Adversary {
    Side adversary;
}
