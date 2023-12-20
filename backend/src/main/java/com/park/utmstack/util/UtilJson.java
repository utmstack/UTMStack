package com.park.utmstack.util;

import com.jayway.jsonpath.JsonPath;
import com.jayway.jsonpath.PathNotFoundException;
import com.jayway.jsonpath.Predicate;

public class UtilJson {
    private static final String CLASSNAME = "UtilJson";

    /**
     * Using the JsonPath lib to read and filter a given Json
     *
     * @param jsonPath The JSON path to read
     * @param json     The JSON string
     * @param filter   Any filter you want to apply
     * @return Could be anything
     */
    public static <T> T read(String jsonPath, String json, Predicate... filter) {
        final String ctx = CLASSNAME + ".getValueFromJsonPath";
        try {
            return JsonPath.parse(json).read(jsonPath, filter);
        } catch (PathNotFoundException e) {
            return null;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}
