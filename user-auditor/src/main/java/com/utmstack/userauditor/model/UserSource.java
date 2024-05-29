package com.utmstack.userauditor.model;

import lombok.Getter;
import lombok.Setter;

import javax.persistence.*;
import java.util.List;

/**
 * A UtmAuditorUserSources.
 */

@Entity
@Table(name = "utm_user_source")
@Getter
@Setter
public class UserSource extends Base {

    @Column(name = "index_pattern")
    private String indexPattern;

    @Column(name = "index_name")
    private String indexName;

    @OneToMany(mappedBy = "source", fetch = FetchType.LAZY, cascade = CascadeType.ALL)
    List<SourceScan> sources;

    @OneToMany(mappedBy = "source", fetch = FetchType.LAZY, cascade = CascadeType.ALL)
    List<SourceFilter> filters;
}
