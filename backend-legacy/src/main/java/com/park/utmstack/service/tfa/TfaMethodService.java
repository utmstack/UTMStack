package com.park.utmstack.service.tfa;

import com.park.utmstack.domain.User;
import com.park.utmstack.domain.tfa.TfaMethod;
import com.park.utmstack.service.dto.tfa.init.TfaInitResponse;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyResponse;

public interface TfaMethodService {
    TfaMethod getMethod();

    TfaInitResponse initiateSetup(User use);

    TfaVerifyResponse verifyCode(User use, String code);

    void persistConfiguration(User use);

    void generateChallenge(User user);

    void regenerateChallenge(User user);

    long expirationTimeSeconds();

}

