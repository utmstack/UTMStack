package com.park.utmstack.util.elastic;

import org.springframework.data.domain.Pageable;

public class SqlPaginationUtil {

    public static String applyPagination(String query, Pageable pageable) {
        String upper = query.toUpperCase();

        boolean hasLimit = upper.contains("LIMIT");
        boolean hasOffset = upper.contains("OFFSET");

        if (hasLimit && hasOffset) {
            return query;
        } else if (hasLimit && !hasOffset) {
            int offset = pageable.getPageNumber() * pageable.getPageSize();
            return query + " OFFSET " + offset;
        } else if (!hasLimit && hasOffset) {
            int pageSize = pageable.getPageSize();
            return query + " LIMIT " + pageSize;
        } else {
            int pageSize = pageable.getPageSize();
            int offset = (pageable.getPageNumber() - 1) * pageSize;
            return query + " LIMIT " + pageSize + " OFFSET " + offset;
        }
    }
}

