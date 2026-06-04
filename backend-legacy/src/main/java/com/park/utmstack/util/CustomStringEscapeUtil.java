package com.park.utmstack.util;

import org.apache.commons.text.translate.AggregateTranslator;
import org.apache.commons.text.translate.LookupTranslator;

import java.util.HashMap;
import java.util.Map;

public class CustomStringEscapeUtil {

    private static final String[] CHARS = new String[] {"+", "-", "=", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "\\", "/"};

    private static final Map<CharSequence, CharSequence> ESCAPE_CUSTOM_MAP = new HashMap<>();

    static {
        ESCAPE_CUSTOM_MAP.put("<", "");
        ESCAPE_CUSTOM_MAP.put(">", "");
        for (String s : CHARS)
            ESCAPE_CUSTOM_MAP.put(s, "\\" + s);
    }

    public static String openSearchQueryStringEscap(String str) {
        return new AggregateTranslator(new LookupTranslator(ESCAPE_CUSTOM_MAP)).translate(str);
    }
}
