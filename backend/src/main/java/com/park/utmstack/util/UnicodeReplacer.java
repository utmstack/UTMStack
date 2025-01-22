package com.park.utmstack.util;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class UnicodeReplacer {
    private static final Pattern UNICODE_PATTERN = Pattern.compile("\\\\u([0-9A-Fa-f]{4})");

    /**
     * Method to replace Unicode sequences with their corresponding characters.
     *
     * @param input The text containing Unicode sequences.
     * @return The text with Unicode sequences replaced by their actual characters.
     */
    public static String replaceUnicode(String input) {
        Matcher matcher = UNICODE_PATTERN.matcher(input);
        StringBuilder result = new StringBuilder();

        while (matcher.find()) {
            // Converts the Unicode sequence to its equivalent character
            String unicode = matcher.group(1);
            char character = (char) Integer.parseInt(unicode, 16);
            matcher.appendReplacement(result, String.valueOf(character));
        }

        matcher.appendTail(result);
        return result.toString();
    }
}