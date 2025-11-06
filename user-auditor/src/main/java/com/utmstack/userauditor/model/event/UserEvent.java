package com.utmstack.userauditor.model.event;

import lombok.Builder;
import lombok.Getter;
import lombok.Setter;
import org.opensearch.client.opensearch._types.aggregations.TopHitsAggregate;

@Builder
@Getter
@Setter
public class UserEvent {

    private final String name;

    private final TopHitsAggregate topEvents;
}
