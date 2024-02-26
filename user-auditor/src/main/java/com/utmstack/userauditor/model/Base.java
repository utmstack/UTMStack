package com.utmstack.userauditor.model;

import com.utmstack.userauditor.model.audit.AuditListener;

import lombok.*;

import javax.persistence.*;
import java.io.Serializable;

@Getter
@Setter
@MappedSuperclass
@EntityListeners(AuditListener.class)

public class Base implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.AUTO)
    private Long id;

    @Column(name = "active", columnDefinition = "boolean default true")
    private boolean active = true;

}
