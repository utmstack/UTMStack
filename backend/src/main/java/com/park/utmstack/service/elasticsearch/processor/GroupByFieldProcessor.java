package com.park.utmstack.service.elasticsearch.processor;

import lombok.RequiredArgsConstructor;

import java.util.*;
import java.util.stream.Collectors;

@RequiredArgsConstructor
public class GroupByFieldProcessor implements SearchResultProcessor {

    private final String fieldName;

    @Override
    public List<Map<String, Object>> process(List<Map<String, Object>> rawResults) {
        Map<Object, Map<String, Object>> byId = new LinkedHashMap<>();
        for (Map<String, Object> item : rawResults) {
            Object id = item.get("id");
            if (id != null) {
                byId.put(id, item);
            }
        }

        Map<Object, List<Map<String, Object>>> childrenByParent = new LinkedHashMap<>();
        Set<Object> childIds = new HashSet<>();

        for (Map<String, Object> item : rawResults) {
            Object parentId = item.get("parentId");
            if (parentId != null && byId.containsKey(parentId)) {
                childrenByParent.computeIfAbsent(parentId, k -> new ArrayList<>()).add(item);
                childIds.add(item.get("id"));
            }
        }

        List<Map<String, Object>> result = new ArrayList<>();
        for (Map<String, Object> item : rawResults) {
            Object id = item.get("id");

            if (id != null && !childIds.contains(id)) {
                Map<String, Object> enriched = new LinkedHashMap<>(item);

                if (childrenByParent.containsKey(id)) {

                    List<Map<String, Object>> children = childrenByParent.get(id);
                    enriched.put("children", processChildren(children, childrenByParent));
                    enriched.put("groupSize", children.size());
                    enriched.put("hasChildren", true);
                } else {

                    enriched.put("children", Collections.emptyList());
                    enriched.put("hasChildren", false);
                }
                result.add(enriched);
            }
        }
        return result;
    }

    private List<Map<String, Object>> processChildren(List<Map<String, Object>> children,
                                                      Map<Object, List<Map<String, Object>>> childrenByParent) {
        List<Map<String, Object>> processedChildren = new ArrayList<>();

        for (Map<String, Object> child : children) {
            Map<String, Object> enrichedChild = new LinkedHashMap<>(child);
            Object childId = child.get("id");

            if (childId != null && childrenByParent.containsKey(childId)) {
                List<Map<String, Object>> grandChildren = childrenByParent.get(childId);
                enrichedChild.put("children", processChildren(grandChildren, childrenByParent));
                enrichedChild.put("groupSize", grandChildren.size());
                enrichedChild.put("hasChildren", true);
            } else {
                enrichedChild.put("children", Collections.emptyList());
                enrichedChild.put("hasChildren", false);
            }
            processedChildren.add(enrichedChild);
        }
        return processedChildren;
    }
}

