package com.park.utmstack.service.tfa;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.service.UtmConfigurationParameterService;
import com.park.utmstack.service.dto.tfa.enroll.TfaEnrollRequest;
import com.park.utmstack.service.dto.tfa.verify.TfaVerifyRequest;
import com.park.utmstack.util.exceptions.TfaVerificationException;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

import static com.park.utmstack.config.Constants.PROP_TFA_METHOD;

@Service
@RequiredArgsConstructor
public class TfaEnrollmentService {

    private final TfaService tfaService;
    private final UtmConfigurationParameterService utmConfigurationParameterService;

    public void enrollTfa(User user, TfaEnrollRequest request) {
        // --- INIT ---
        if (request.getInitMethod() != null) {
            tfaService.initiateSetup(user, request.getInitMethod());
        }

        // --- VERIFY ---
        if (request.getVerifyMethod() != null && request.getVerifyCode() != null) {
            var verifyRequest = new TfaVerifyRequest(request.getVerifyMethod(), request.getVerifyCode());
            var verifyResponse = tfaService.verifyCode(user, verifyRequest);
            if (!verifyResponse.isValid()) {
                throw new TfaVerificationException("Código TFA inválido: " + verifyResponse.getMessage());
            }
        }

        // --- COMPLETE ---
        if (request.getCompleteRequest() != null) {
            List<UtmConfigurationParameter> tfaParams = utmConfigurationParameterService.getConfigParameterBySectionId(Constants.TFA_SETTING_ID);
            var complete = request.getCompleteRequest();
            for (UtmConfigurationParameter param : tfaParams) {
                switch (param.getConfParamShort()) {
                    case PROP_TFA_METHOD -> param.setConfParamValue(String.valueOf(complete.getMethod()));
                    case Constants.PROP_TFA_ENABLE -> param.setConfParamValue(String.valueOf(complete.isEnable()));
                }
            }
            tfaService.persistConfiguration(complete.getMethod());
            utmConfigurationParameterService.saveAllConfigParams(tfaParams);
            tfaService.generateChallenge(user);
        }
    }
}