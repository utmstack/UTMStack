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

    @NotNull
    private Integer fetchSize;
}
