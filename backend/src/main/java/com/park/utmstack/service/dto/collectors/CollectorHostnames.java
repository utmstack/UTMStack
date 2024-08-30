package com.park.utmstack.service.dto.collectors;

import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
@NoArgsConstructor
public class CollectorHostnames {
    public List<String> hostname = new ArrayList<String>();
}
