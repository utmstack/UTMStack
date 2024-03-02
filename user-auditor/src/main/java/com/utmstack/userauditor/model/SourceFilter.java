package com.utmstack.userauditor.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import lombok.Getter;
import lombok.Setter;

import javax.persistence.*;

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
