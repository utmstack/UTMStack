package com.park.utmstack.service.tfa;

import com.park.utmstack.domain.tfa.TfaMethod;
import org.springframework.stereotype.Service;

@Service
public class ConfigService {
    public void enableTfa(String username, TfaMethod method, String secret) {
        // Persistir en base de datos: método + secret
    }

    public void disableTfa(String username) {
        // Eliminar configuración TFA del usuario
    }
}

