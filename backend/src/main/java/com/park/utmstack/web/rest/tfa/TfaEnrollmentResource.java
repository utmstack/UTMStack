package com.park.utmstack.web.rest.tfa;

import com.park.utmstack.domain.User;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.tfa.enroll.TfaEnrollRequest;
import com.park.utmstack.service.tfa.TfaEnrollmentService;
import io.swagger.v3.oas.annotations.Hidden;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequiredArgsConstructor
@Slf4j
@Hidden
@RequestMapping("api/tfa")
public class TfaEnrollmentResource {

    private static final String CLASSNAME = "TfaEnrollmentController";

    private final UserService userService;
    private final TfaEnrollmentService tfaEnrollmentService;


    @PostMapping("/enroll")
    public ResponseEntity<?> enrollTfa(@RequestBody TfaEnrollRequest request) {
        final String ctx = CLASSNAME + ".enrollTfa";
        User user = userService.getCurrentUserLogin();

        tfaEnrollmentService.enrollTfa(user, request);
        return ResponseEntity.ok().build();
    }
}

