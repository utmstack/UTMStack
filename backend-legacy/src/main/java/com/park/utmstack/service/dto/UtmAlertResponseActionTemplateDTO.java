package com.park.utmstack.service.dto;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Size;
import java.io.Serializable;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class UtmAlertResponseActionTemplateDTO implements Serializable {

    private Long id;

    @NotBlank
    @Size(max = 150)
    private String title;

    @Size(max = 512)
    private String description;

    @NotBlank
    private String command;

    private Boolean systemOwner;

    public static UtmAlertResponseActionTemplateDTO fromEntity(UtmAlertResponseActionTemplate entity) {
        return UtmAlertResponseActionTemplateDTO.builder()
                .id(entity.getId())
                .title(entity.getTitle())
                .description(entity.getDescription())
                .systemOwner(entity.getSystemOwner())
                .command(entity.getCommand())
                .build();
    }
}
