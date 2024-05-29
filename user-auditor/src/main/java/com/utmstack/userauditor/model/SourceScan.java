package com.utmstack.userauditor.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.utmstack.userauditor.model.audit.Auditable;
import lombok.*;

import javax.persistence.*;

import java.time.LocalDateTime;

@Entity
@Table(name = "utm_source_scan")
@Getter
@Setter
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class SourceScan extends Base implements Auditable {

    @Column(name = "next_execution_date")
    private LocalDateTime executionDate;

    @ManyToOne
    @JoinColumn(name = "user_sources_id")
    @JsonIgnore
    UserSource source;

    @Embedded
    Audit audit;
}
