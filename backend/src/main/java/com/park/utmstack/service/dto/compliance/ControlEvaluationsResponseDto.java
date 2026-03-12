package com.park.utmstack.service.dto.compliance;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDate;
import java.util.List;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ControlEvaluationsResponseDto {
    LocalDate startDate;
    LocalDate endDate;
    List<UtmComplianceControlEvaluationsGroupedDto> evaluations;
}