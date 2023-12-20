package com.park.utmstack.service.dto.incident;

import javax.validation.constraints.NotNull;

public class NewIncidentNoteDTO {
    @NotNull
    Long incidentId;

    @NotNull
    String noteText;

    public NewIncidentNoteDTO() {
    }

    public Long getIncidentId() {
        return incidentId;
    }

    public void setIncidentId(Long incidentId) {
        this.incidentId = incidentId;
    }

    public String getNoteText() {
        return noteText;
    }

    public void setNoteText(String noteText) {
        this.noteText = noteText;
    }
}
