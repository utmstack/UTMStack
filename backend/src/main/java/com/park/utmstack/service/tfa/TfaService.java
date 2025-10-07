package com.park.utmstack.service.tfa;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.tfa.init.TfaInitResponse;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyRequest;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class TfaService {

    private final List<TfaMethodService> methodServices;
    private final UserService userService;

    private TfaMethodService getMethodService(TfaMethod method) {
        return methodServices.stream()
                .filter(service -> service.getMethod().equals(method))
                .findFirst()
                .orElseThrow(() -> new IllegalArgumentException("TFA method not supported: " + method));
    }

    public TfaInitResponse initiateSetup(User user, TfaMethod method) {
        TfaMethodService selected = getMethodService(method);
        return selected.initiateSetup(user);
    }

    public TfaVerifyResponse verifyCode(User user, TfaVerifyRequest request) {
        TfaMethodService selected = getMethodService(request.getMethod());
        return selected.verifyCode(user, request.getCode());
    }

    public void persistConfiguration(TfaMethod method) {
        User user = userService.getCurrentUserLogin();
        TfaMethodService selected = getMethodService(method);
        selected.persistConfiguration(user);
    }

    public long generateChallenge(User user) {

        TfaMethod method = TfaMethod.valueOf(user.getTfaMethod());

        TfaMethodService selected = getMethodService(method);
        selected.generateChallenge(user);

        return selected.expirationTimeSeconds();
    }

    public void regenerateChallenge(User user) {

        TfaMethod method = TfaMethod.valueOf(user.getTfaMethod());

        TfaMethodService selected = getMethodService(method);
        selected.regenerateChallenge(user);
    }
}

