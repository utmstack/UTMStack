package com.park.utmstack.util;

import com.park.utmstack.util.exceptions.UtmPageNumberNotSupported;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.http.HttpHeaders;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.List;
import java.util.stream.Collectors;


public class UtilPagination {
    public static int getFirst(int pageSize, int pageNumber) throws UtmPageNumberNotSupported {
        String ctx = "UtilPagination.getFirst";
        if (pageNumber <= 0)
            throw new UtmPageNumberNotSupported(ctx + ": Invalid page number: " + pageNumber);
        return pageSize * pageNumber - (pageSize - 1);
    }

    public static int getFirstForElasticsearch(int pageSize, int pageNumber) {
        if (pageNumber <= 0)
            pageNumber = 1;
        return pageSize * pageNumber - (pageSize);
    }

    public static int getFirstForNativeSql(int pageSize, int pageNumber) {
        return pageSize * pageNumber;
    }

    public static <T> HttpHeaders generatePaginationHttpHeaders(Long totalElements, int pageNumber, int pageSize,
                                                                String baseUrl) {
        long totalPages = getTotalPages(totalElements, pageSize);
        HttpHeaders headers = new HttpHeaders();
        headers.add("X-Total-Count", String.valueOf(totalElements));
        String link = "";
        if ((pageNumber + 1) < totalPages) {
            link = "<" + generateUri(baseUrl, pageNumber + 1, pageSize) + ">; rel=\"next\",";
        }
        // prev link
        if ((pageNumber) > 0) {
            link += "<" + generateUri(baseUrl, pageNumber - 1, pageSize) + ">; rel=\"prev\",";
        }

        // last and first link
        long lastPage = 0;
        if (totalPages > 0) {
            lastPage = totalPages - 1;
        }

        link += "<" + generateUri(baseUrl, lastPage, pageSize) + ">; rel=\"last\",";
        link += "<" + generateUri(baseUrl, 0, pageSize) + ">; rel=\"first\"";
        headers.add(HttpHeaders.LINK, link);
        return headers;
    }

    private static String generateUri(String baseUrl, long page, long size) {
        return UriComponentsBuilder.fromUriString(baseUrl).queryParam("page", page).queryParam("size", size).toUriString();
    }

    private static long getTotalPages(Long totalElements, int pageSize) {
        long result = totalElements / pageSize;
        long rest = totalElements % pageSize;

        if (rest != 0)
            return result + 1;
        return result;
    }

    /**
     * Add pagination to a native query
     *
     * @param query    A native query to add pagination
     * @param pageable Pagination info
     * @return A string with the paginated query
     */
    public static String paginateAndSortNativeSqlQuery(String query, Pageable pageable) {
        StringBuilder sb = new StringBuilder(query);

        Sort sort = pageable.getSort();

        if (sort.isSorted()) {
            sb.append(" ORDER BY ");
            boolean firstProperty = true;

            List<Sort.Order> orders = sort.stream().collect(Collectors.toList());

            for (Sort.Order order : orders) {
                sb.append(String.format(firstProperty ? "%1$s %2$s" : ", %1$s %2$s", order.getProperty(), order.getDirection().name()));
                firstProperty = false;
            }
        }

        if (pageable.isPaged())
            sb.append(String.format(" OFFSET %1$s LIMIT %2$s", pageable.getOffset(), pageable.getPageSize()));

        return sb.toString();
    }
}
