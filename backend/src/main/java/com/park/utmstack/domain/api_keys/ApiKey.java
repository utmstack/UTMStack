package com.park.utmstack.domain.api_keys;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.persistence.*;
import java.io.Serializable;
import java.time.Instant;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "api_keys")
public class ApiKey implements Serializable {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private Long userId;

    @Column(nullable = false)
    private String name;

    @Column(nullable = false)
    private String apiKey;

    @Column
    private String allowedIp;

    @Column(nullable = false)
    private Instant createdAt;

    private Instant generatedAt;

    @Column
    private Instant expiresAt;
}
