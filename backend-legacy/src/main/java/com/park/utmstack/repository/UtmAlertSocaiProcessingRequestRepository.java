package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmAlertSocaiProcessingRequest;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@SuppressWarnings("unused")
@Repository
public interface UtmAlertSocaiProcessingRequestRepository extends JpaRepository<UtmAlertSocaiProcessingRequest, String> {

}
