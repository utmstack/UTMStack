package com.park.utmstack.util;

import com.park.utmstack.config.Constants;

import java.time.DateTimeException;
import java.time.ZoneId;

import static com.park.utmstack.config.Constants.PROP_DATE_SETTINGS_TIMEZONE;

public class TimezoneUtil {

    public static ZoneId getAppTimezone() {
        ZoneId timezone;
        try {
            timezone = ZoneId.of(Constants.CFG.get(PROP_DATE_SETTINGS_TIMEZONE));
        } catch (DateTimeException e) {
            timezone = ZoneId.systemDefault();
        }
        return timezone;
    }
}
