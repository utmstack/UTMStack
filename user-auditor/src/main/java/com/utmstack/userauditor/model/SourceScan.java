package com.utmstack.userauditor.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.utmstack.userauditor.model.audit.Auditable;
import lombok.*;

import jakarta.persistence.*;

import java.time.LocalDate;

@Entity
@Table(name = "utm_source_scan")
@Getter
@Setter
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class SourceScan extends Base implements Auditable {

    @Column(name = "last_execution_date")
    private LocalDate executionDate;

    @ManyToOne
    @JoinColumn(name = "user_sources_id")
    @JsonIgnore
    UserSource source;

    @Embedded
    Audit audit;
}
