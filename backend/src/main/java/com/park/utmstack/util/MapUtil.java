package com.park.utmstack.util;

import org.springframework.lang.Nullable;
import org.springframework.util.Assert;
import org.springframework.util.StringUtils;

import java.util.Iterator;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.function.Function;
import java.util.function.UnaryOperator;
import java.util.stream.Collectors;
import java.util.stream.Stream;

public class MapUtil {
    /**
     * Flatten a hierarchical {@link Map} into a flat {@link Map} with key names using property dot notation.
     *
     * @param inputMap
     *     must not be {@literal null}.
     *
     * @return the resulting {@link Map}.
     */
    public static Map<String, Object> flatten(Map<String, ?> inputMap) {

        Assert.notNull(inputMap, "Input Map must not be null");

        Map<String, Object> resultMap = new LinkedHashMap<>();

        doFlatten("", inputMap.entrySet().iterator(), resultMap, UnaryOperator.identity(), false);

        return resultMap;
    }

    /**
     * Flatten a hierarchical {@link Map} into a flat {@link Map} with key names using property dot notation.
     *
     * @param inputMap
     *     must not be {@literal null}.
     *
     * @return the resulting {@link Map}.
     *
     * @since 2.0
     */
    public static Map<String, String> flattenToStringMap(Map<String, ?> inputMap, boolean collectionsAsCommaSeparated) {

        Assert.notNull(inputMap, "Input Map must not be null");

        Map<String, String> resultMap = new LinkedHashMap<>();

        doFlatten("", inputMap.entrySet().iterator(), resultMap, it -> it == null ? null : it.toString(),
                  collectionsAsCommaSeparated);

        return resultMap;
    }

    private static void doFlatten(String propertyPrefix, Iterator<? extends Map.Entry<String, ?>> inputMap,
                                  Map<String, ?> resultMap, Function<Object, Object> valueTransformer,
                                  boolean collectionsAsCommaSeparated) {
        if (StringUtils.hasText(propertyPrefix))
            propertyPrefix = propertyPrefix + ".";

        while (inputMap.hasNext()) {
            Map.Entry<String, ?> entry = inputMap.next();
            flattenElement(propertyPrefix.concat(entry.getKey()), entry.getValue(), resultMap, valueTransformer,
                           collectionsAsCommaSeparated);
        }
    }

    @SuppressWarnings("unchecked")
    private static void flattenElement(String propertyPrefix, @Nullable Object source, Map<String, ?> resultMap,
                                       Function<Object, Object> valueTransformer, boolean collectionsAsCommaSeparated) {

        if (source instanceof Iterable) {
            flattenCollection(propertyPrefix, (Iterable<Object>) source, resultMap, valueTransformer,
                              collectionsAsCommaSeparated);
            return;
        }

        if (source instanceof Map) {
            doFlatten(propertyPrefix, ((Map<String, ?>) source).entrySet().iterator(), resultMap, valueTransformer,
                      collectionsAsCommaSeparated);
            return;
        }
        ((Map) resultMap).put(propertyPrefix, valueTransformer.apply(source));
    }

    private static void flattenCollection(String propertyPrefix, Iterable<Object> iterable, Map<String, ?> resultMap,
                                          Function<Object, Object> valueTransformer, boolean collectionsAsCommaSeparated) {
        if (collectionsAsCommaSeparated) {
            flattenElement(propertyPrefix, Stream.of(iterable).map(String::valueOf).collect(Collectors.joining(","))
                                                 .replace("[", "").replace("]", ""), resultMap, valueTransformer, true);
        } else {
            int counter = 0;
            for (Object element : iterable) {
                flattenElement(propertyPrefix + "[" + counter + "]", element, resultMap, valueTransformer, false);
                counter++;
            }
        }
    }
}
