package com.park.utmstack.service.dto.elastic;

import com.park.utmstack.validation.elasticsearch.SqlSelectOnly;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotNull;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class SqlSearchDto {

    @SqlSelectOnly
    private String query;

    /**
     * Specifies the maximum number of results to fetch per query execution.
     * Acceptable values are positive integers; if null or not set, the default fetch size will be used.
     * This parameter can be used to limit the number of records returned by the SQL query.
     */
    private Integer fetchSize;
}
