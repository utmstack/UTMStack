package com.utmstack.userauditor.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.utmstack.userauditor.model.audit.Auditable;
import lombok.*;

import javax.persistence.*;

/**
 * A UtmAuditorUserAttributes.
 */
@Entity
@Table(name = "utm_user_attribute")
@Getter
@Setter
@Builder
@AllArgsConstructor
@RequiredArgsConstructor
public class UserAttribute extends Base implements Auditable {

    @Column(name = "attribute_key")
    private String attributeKey;

    @Column(name = "attribute_value")
    private String attributeValue;

    @Embedded
    Audit audit;

    @ManyToOne
    @JoinColumn(name = "user_id")
    @JsonIgnore
    private User user;
}
