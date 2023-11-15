package com.utmstack.userauditor.model.winevent;

import com.utmstack.userauditor.model.UserAttribute;
import lombok.*;

import java.util.List;
import java.util.Map;


@NoArgsConstructor
@AllArgsConstructor
@Getter
@Setter
@Builder

public class UserLog {
   Map<String, List<UserAttribute>> users;
}
