package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncidentNote;
import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;
import com.park.utmstack.repository.incident.UtmIncidentNoteRepository;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.incident.NewIncidentNoteDTO;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentNote.
 */
@Service
@Transactional
public class UtmIncidentNoteService {
    private final String CLASSNAME = "UtmIncidentNoteService";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentNoteService.class);

    private final UtmIncidentNoteRepository utmIncidentNoteRepository;

    private final UtmIncidentHistoryService utmIncidentHistoryService;

    private final UserService userService;

    public UtmIncidentNoteService(UtmIncidentNoteRepository utmIncidentNoteRepository, UtmIncidentHistoryService utmIncidentHistoryService, UserService userService) {
        this.utmIncidentNoteRepository = utmIncidentNoteRepository;
        this.utmIncidentHistoryService = utmIncidentHistoryService;
        this.userService = userService;
    }

    /**
     * Save a utmIncidentNote.
     *
     * @param newIncidentNoteDTO the entity to save
     * @return the persisted entity
     */
    public UtmIncidentNote save(NewIncidentNoteDTO newIncidentNoteDTO, Boolean edit) {
        final String ctx = CLASSNAME + ".createHistory";
        try {
            log.debug("Request to save UtmIncidentNote : {}", newIncidentNoteDTO);
            UtmIncidentNote newNote = new UtmIncidentNote();
            newNote.setIncidentId(newIncidentNoteDTO.getIncidentId());
            newNote.setNoteText(newIncidentNoteDTO.getNoteText());
            newNote.setNoteSendDate(Instant.now());
            String userLogin = userService.getCurrentUserLogin().getLogin();
            newNote.setNoteSendBy(userLogin);
            UtmIncidentNote note = utmIncidentNoteRepository.save(newNote);
            utmIncidentHistoryService.createHistory(edit ? IncidentHistoryActionEnum.INCIDENT_NOTE_CHANGE : IncidentHistoryActionEnum.INCIDENT_NOTE_ADD, note.getIncidentId(),
                edit ? "Note changed" : "Note added", "New note added to incident");
            return note;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all the utmIncidentNotes.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentNote> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIncidentNotes");
        return utmIncidentNoteRepository.findAll(pageable);
    }


    /**
     * Get one utmIncidentNote by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncidentNote> findOne(Long id) {
        log.debug("Request to get UtmIncidentNote : {}", id);
        return utmIncidentNoteRepository.findById(id);
    }

    /**
     * Delete the utmIncidentNote by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentNote : {}", id);
        utmIncidentNoteRepository.deleteById(id);
    }
}
