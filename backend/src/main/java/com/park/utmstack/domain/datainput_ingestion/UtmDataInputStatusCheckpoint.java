package com.park.utmstack.domain.datainput_ingestion;

import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import javax.persistence.*;
import java.time.Instant;

@Entity
@Getter
@Setter
@NoArgsConstructor
@Table(name = "utm_data_input_status_checkpoint")
public class UtmDataInputStatusCheckpoint {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "last_processed_timestamp", nullable = false)
    private Instant lastProcessedTimestamp;
}
