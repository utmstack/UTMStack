package com.park.utmstack.service.dto.incident;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmIncidentNote entity. This class is used in UtmIncidentNoteResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-incident-notes?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIncidentNoteCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter incidentId;

    private StringFilter noteText;

    private InstantFilter noteSendDate;

    private StringFilter noteSendBy;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getIncidentId() {
        return incidentId;
    }

    public void setIncidentId(LongFilter incidentId) {
        this.incidentId = incidentId;
    }

    public StringFilter getNoteText() {
        return noteText;
    }

    public void setNoteText(StringFilter noteText) {
        this.noteText = noteText;
    }

    public InstantFilter getNoteSendDate() {
        return noteSendDate;
    }

    public void setNoteSendDate(InstantFilter noteSendDate) {
        this.noteSendDate = noteSendDate;
    }

    public StringFilter getNoteSendBy() {
        return noteSendBy;
    }

    public void setNoteSendBy(StringFilter noteSendBy) {
        this.noteSendBy = noteSendBy;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmIncidentNoteCriteria that = (UtmIncidentNoteCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(incidentId, that.incidentId) &&
            Objects.equals(noteText, that.noteText) &&
            Objects.equals(noteSendDate, that.noteSendDate) &&
            Objects.equals(noteSendBy, that.noteSendBy);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        incidentId,
        noteText,
        noteSendDate,
        noteSendBy
        );
    }

    @Override
    public String toString() {
        return "UtmIncidentNoteCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (incidentId != null ? "incidentId=" + incidentId + ", " : "") +
                (noteText != null ? "noteText=" + noteText + ", " : "") +
                (noteSendDate != null ? "noteSendDate=" + noteSendDate + ", " : "") +
                (noteSendBy != null ? "noteSendBy=" + noteSendBy + ", " : "") +
            "}";
    }

}
