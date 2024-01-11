package com.park.utmstack.util;

import org.apache.commons.text.translate.AggregateTranslator;
import org.apache.commons.text.translate.LookupTranslator;

import java.util.HashMap;
import java.util.Map;

public class CustomStringEscapeUtil {
    public static String opensearchQueryStringEscape(String str) {
        final String[] chars = new String[] {"+", "-", "=", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "\\", "/"};
        final Map<CharSequence, CharSequence> escapeCustomMap = new HashMap<>();
        escapeCustomMap.put("<", "");
        escapeCustomMap.put(">", "");
        for (String s : chars)
            escapeCustomMap.put(s, "\\\\" + s);
        return new AggregateTranslator(new LookupTranslator(escapeCustomMap)).translate(str);
    }
}
