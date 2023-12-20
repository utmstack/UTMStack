package com.park.utmstack.web.rest.agent_manager;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.agent_manager.AgentGroupGrpcService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.agent_manager.AgentGroupDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentGroupsDTO;
import com.park.utmstack.service.grpc.AgentGroup;
import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.application_modules.UtmModuleResource;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.vm.AgentGroupVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;

@RestController
@RequestMapping("/api")
public class AgentManagerGroupResource {
    private static final String CLASSNAME = "AgentManagerGroupResource";
    private static final String ENTITY_NAME = "AgentManagerGroup";
    private final Logger log = LoggerFactory.getLogger(UtmModuleResource.class);
    private final AgentGroupGrpcService agentGroupGrpcService;
    private final ApplicationEventService eventService;

    public AgentManagerGroupResource(AgentGroupGrpcService agentGroupGrpcService,
                                     ApplicationEventService eventService) {
        this.agentGroupGrpcService = agentGroupGrpcService;
        this.eventService = eventService;
    }

    @GetMapping("/agent-manager-group/get-by-filter")
    public ResponseEntity<List<AgentGroupDTO>> listAgentGroups(
        @RequestParam(required = false) Integer pageNumber,
        @RequestParam(required = false) Integer pageSize,
        @RequestParam(required = false) String searchQuery,
        @RequestParam(required = false) String sortBy) {

        final String ctx = CLASSNAME + ".listAgentGroups";
        try {
            ListRequest request = ListRequest.newBuilder()
                .setPageNumber(pageNumber != null ? pageNumber : 0)
                .setPageSize(pageSize != null ? pageSize : 0)
                .setSearchQuery(searchQuery != null ? searchQuery : "")
                .setSortBy(sortBy != null ? sortBy : "")
                .build();
            ListAgentGroupsDTO response = agentGroupGrpcService.listGroups(request);
            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", Long.toString(response.getTotal()));
            return ResponseEntity.ok().headers(headers).body(response.getGroups());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }


    /**
     * POST  /agent-manager-group : Create a new agentGroup.
     *
     * @param agentGroup the agentGroup to create
     * @return the ResponseEntity with status 201 (Created) and with body the new agentGroup, or with status 400 (Bad Request) if the agentGroup has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/agent-manager-group")
    public ResponseEntity<AgentGroupDTO> createAgentGroupDTO(@Valid @RequestBody AgentGroupVM agentGroup) throws Exception {
        final String ctx = CLASSNAME + ".createAgentGroupDTO";
        try {
            if (agentGroup.getId() != null) {
                throw new BadRequestAlertException("A new agentGroup cannot already have an ID", ENTITY_NAME, "idexists");
            }
            AgentGroupDTO result = agentGroupGrpcService.createGroup(convertToGrpc(agentGroup));
            return ResponseEntity.created(new URI("/api/agent-manager-group/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, String.valueOf(result.getId())))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * PUT  /agent-manager-group : Updates an existing agentGroup.
     *
     * @param agentGroup the agentGroup to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated agentGroup,
     * or with status 400 (Bad Request) if the agentGroup is not valid,
     * or with status 500 (Internal Server Error) if the agentGroup couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/agent-manager-group")
    public ResponseEntity<AgentGroupDTO> updateAgentGroupDTO(@Valid @RequestBody AgentGroupVM agentGroup) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateAgentGroupDTO";
        try {
            if (agentGroup.getId() == null) {
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
            }
            AgentGroupDTO result = agentGroupGrpcService.editGroup(convertToGrpc(agentGroup));
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, agentGroup.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * DELETE  /agent-manager-group/:id : delete the "id" agentGroup.
     *
     * @param id the id of the agentGroup to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/agent-manager-group/{id}")
    public ResponseEntity<Void> deleteAgentGroupDTO(@PathVariable String id) {
        final String ctx = CLASSNAME + ".updateAgentGroupDTO";
        try {
            agentGroupGrpcService.deleteGroup(Long.valueOf(id));
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id)).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    private AgentGroup convertToGrpc(AgentGroupVM agentGroupVM) {
        return AgentGroup.newBuilder()
            .setGroupDescription(agentGroupVM.getGroupDescription())
            .setGroupName(agentGroupVM.getGroupName())
            .setId(Math.toIntExact(agentGroupVM.getId()))
            .build();
    }
}
