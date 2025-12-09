package com.park.utmstack.util.elastic;

import org.springframework.data.domain.Pageable;

public class SqlPaginationUtil {

    public static String applyPagination(String query, Pageable pageable) {
        String upper = query.toUpperCase();

        boolean hasLimit = upper.contains("LIMIT");
        boolean hasOffset = upper.contains("OFFSET");

        if (hasLimit && hasOffset) {
            return query;
        } else if (hasLimit) {
            int offset = (pageable.getPageNumber() -1 ) * pageable.getPageSize();
            return query + " OFFSET " + offset;
        } else if (hasOffset) {
            int pageSize = pageable.getPageSize();
            return query + " LIMIT " + pageSize;
        } else {
            int pageSize = pageable.getPageSize();
            int offset = (pageable.getPageNumber() - 1) * pageSize;
            return query + " LIMIT " + pageSize + " OFFSET " + offset;
        }
    }
}

