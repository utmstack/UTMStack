package com.park.utmstack.repository.log_analyzer;

import com.park.utmstack.domain.log_analyzer.LogAnalyzerQuery;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the LogAnalyzerQuery entity.
 */
@SuppressWarnings("unused")
@Repository
public interface LogAnalyzerQueryRepository extends JpaRepository<LogAnalyzerQuery, Long>, JpaSpecificationExecutor<LogAnalyzerQuery> {

}
