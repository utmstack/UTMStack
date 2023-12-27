package com.park.utmstack.domain.logstash_pipeline.enums;

import java.util.Arrays;
import java.util.Optional;

/**
 * Supported input variants
 */
public enum InputPlugin {
    TCP,
    TCP_SSL,
    UDP,
    UDP_SSL,
    HTTP,
    HTTP_SSL;

    public static InputPlugin getPluginEnum(String proto) {
        Optional<InputPlugin> val = Arrays
            .stream(InputPlugin.values())
            .filter(repVal -> (repVal.name()).
                compareToIgnoreCase(proto) == 0)
            .findFirst();
        return val.get();
    }
}
