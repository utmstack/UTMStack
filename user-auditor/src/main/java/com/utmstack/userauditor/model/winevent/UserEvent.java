package com.utmstack.userauditor.model.winevent;

import lombok.Builder;
import lombok.Getter;
import lombok.Setter;
import org.opensearch.client.opensearch._types.aggregations.TopHitsAggregate;

import java.util.List;

@Builder
@Getter
@Setter
public class UserEvent {

    private final String name;

    private final TopHitsAggregate topEvents;
}
