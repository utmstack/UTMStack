package com.park.utmstack.repository.datainput_ingestion;

import com.park.utmstack.domain.datainput_ingestion.UtmDataInputStatusCheckpoint;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface UtmDataInputStatusCheckpointRepository extends JpaRepository<UtmDataInputStatusCheckpoint, Long> {}
