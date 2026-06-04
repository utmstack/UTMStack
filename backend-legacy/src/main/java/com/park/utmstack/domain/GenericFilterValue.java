package com.park.utmstack.domain;

import java.io.Serializable;

public class GenericFilterValue  implements Serializable {

    private String id;
    private String name;
    private Long count;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Long getCount() {
        return count;
    }

    public void setCount(Long count) {
        this.count = count;
    }
}
