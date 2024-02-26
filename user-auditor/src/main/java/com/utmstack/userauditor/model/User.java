package com.utmstack.userauditor.model;

import com.utmstack.userauditor.model.audit.Auditable;
import lombok.*;

import javax.persistence.*;
import java.util.List;

/**
 * A UtmAuditorUsers.
 */
@Entity
@Table(name = "utm_user")
@Getter
@Setter
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class User extends Base implements Auditable {

    @Column(name = "name")
    public String name;

    @Column(name = "sid")
    public String sid;

    @Embedded
    Audit audit;

    @ManyToOne
    @JoinColumn (name = "user_sources_id")
    public UserSource source;

    @OneToMany(mappedBy = "user", fetch = FetchType.LAZY, cascade = CascadeType.ALL)
    List<UserAttribute> attributes;
}
