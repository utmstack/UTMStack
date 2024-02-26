package com.utmstack.userauditor.model;

import lombok.Getter;
import lombok.Setter;

import javax.persistence.*;

import javax.persistence.Embeddable;
import java.time.LocalDateTime;

@Embeddable
@Getter
@Setter
public class Audit {

    @Column(name = "created_date")
    private LocalDateTime createdDate;

    @Column(name = "modified_date")
    private LocalDateTime  modifiedDate;

}
