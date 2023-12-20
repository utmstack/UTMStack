package com.park.utmstack.domain.incident;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.Objects;

/**
 * A UtmIncidentNote.
 */
@Entity
@Table(name = "utm_incident_note")
public class UtmIncidentNote implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "incident_id", nullable = false)
    private Long incidentId;

    @NotNull
    @Size(max = 1000)
    @Column(name = "note_text", length = 1000, nullable = false)
    private String noteText;

    @NotNull
    @Column(name = "note_send_date", nullable = false)
    private Instant noteSendDate;

    @NotNull
    @Column(name = "note_send_by", nullable = false)
    private String noteSendBy;

    // jhipster-needle-entity-add-field - JHipster will add fields here, do not remove
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getIncidentId() {
        return incidentId;
    }

    public UtmIncidentNote incidentId(Long incidentId) {
        this.incidentId = incidentId;
        return this;
    }

    public void setIncidentId(Long incidentId) {
        this.incidentId = incidentId;
    }

    public String getNoteText() {
        return noteText;
    }

    public UtmIncidentNote noteText(String noteText) {
        this.noteText = noteText;
        return this;
    }

    public void setNoteText(String noteText) {
        this.noteText = noteText;
    }

    public Instant getNoteSendDate() {
        return noteSendDate;
    }

    public UtmIncidentNote noteSendDate(Instant noteSendDate) {
        this.noteSendDate = noteSendDate;
        return this;
    }

    public void setNoteSendDate(Instant noteSendDate) {
        this.noteSendDate = noteSendDate;
    }

    public String getNoteSendBy() {
        return noteSendBy;
    }

    public UtmIncidentNote noteSendBy(String noteSendBy) {
        this.noteSendBy = noteSendBy;
        return this;
    }

    public void setNoteSendBy(String noteSendBy) {
        this.noteSendBy = noteSendBy;
    }
    // jhipster-needle-entity-add-getters-setters - JHipster will add getters and setters here, do not remove

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        UtmIncidentNote utmIncidentNote = (UtmIncidentNote) o;
        if (utmIncidentNote.getId() == null || getId() == null) {
            return false;
        }
        return Objects.equals(getId(), utmIncidentNote.getId());
    }

    @Override
    public int hashCode() {
        return Objects.hashCode(getId());
    }

    @Override
    public String toString() {
        return "UtmIncidentNote{" +
            "id=" + getId() +
            ", incidentId=" + getIncidentId() +
            ", noteText='" + getNoteText() + "'" +
            ", noteSendDate='" + getNoteSendDate() + "'" +
            ", noteSendBy='" + getNoteSendBy() + "'" +
            "}";
    }
}
