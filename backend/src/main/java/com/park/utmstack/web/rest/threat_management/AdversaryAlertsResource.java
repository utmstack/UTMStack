package com.park.utmstack.web.rest.threat_management;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.service.dto.threat_management.AdversaryAlertsResponseDto;
import com.park.utmstack.service.threat_management.AdversaryAlertsService;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;

@RestController
@RequiredArgsConstructor
@Slf4j
@Hidden
@RequestMapping("api/adversary")
public class AdversaryAlertsResource {

    private final AdversaryAlertsService adversaryAlertsService;

    @PostMapping("/alerts")
    public ResponseEntity<List<AdversaryAlertsResponseDto>> search(@RequestBody(required = false) List<FilterType> filters) {

            List<AdversaryAlertsResponseDto> responseDto = adversaryAlertsService.fetchAdversaryAlerts(filters);

            if (responseDto.isEmpty())
                return ResponseEntity.status(HttpStatus.NO_CONTENT).body(Collections.emptyList());

            return ResponseEntity.ok().body(responseDto);
    }
}
