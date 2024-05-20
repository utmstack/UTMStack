package com.park.utmstack.service.dto.collectors.dto;

import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;

import java.util.List;

public class ListCollectorsResponseDTO {
    private List<CollectorDTO> collectors;
    private int total;

    public List<CollectorDTO> getCollectors() {
        return collectors;
    }

    public void setCollectors(List<CollectorDTO> collectors) {
        this.collectors = collectors;
    }

    public int getTotal() {
        return total;
    }

    public void setTotal(int total) {
        this.total = total;
    }
}
