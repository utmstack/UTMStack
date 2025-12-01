package com.park.utmstack.web.rest.app_info;

import com.park.utmstack.service.app_info.AppInfoService;
import com.park.utmstack.service.dto.app_info.AppInfoDto;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/info")
@RequiredArgsConstructor
public class AppInfoResource {

    private final AppInfoService appInfoService;

    @GetMapping("/version")
    public ResponseEntity<AppInfoDto> getInfo() throws Exception {
        return ResponseEntity.ok(appInfoService.loadVersionInfo());
    }
}
