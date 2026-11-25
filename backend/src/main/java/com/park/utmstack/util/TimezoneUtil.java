package com.park.utmstack.util;

import com.park.utmstack.config.Constants;

import java.time.DateTimeException;
import java.time.ZoneId;

import static com.park.utmstack.config.Constants.PROP_DATE_SETTINGS_TIMEZONE;

public class TimezoneUtil {

    public static ZoneId getAppTimezone() {
        String tz = Constants.CFG.get(PROP_DATE_SETTINGS_TIMEZONE);
        if (tz == null || tz.isBlank()) {
            return ZoneId.systemDefault();
        }
        try {
            return ZoneId.of(tz);
        } catch (DateTimeException e) {
            return ZoneId.systemDefault();
        }
    }
}
