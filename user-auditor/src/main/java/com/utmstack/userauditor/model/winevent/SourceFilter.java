package com.utmstack.userauditor.model.winevent;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.utmstack.userauditor.model.Base;
import com.utmstack.userauditor.model.UserSource;
import lombok.Getter;
import lombok.Setter;

import jakarta.persistence.*;

/**
 * A UtmAuditorUserSourcesQuery.
 */
@Getter
@Setter
@Entity
@Table(name = "utm_source_filter")
public class SourceFilter extends Base {

    @ManyToOne
    @JoinColumn(name = "user_sources_id")
    @JsonIgnore
    private UserSource source;

    private String field;

    private String value;

    private int operator;

}
