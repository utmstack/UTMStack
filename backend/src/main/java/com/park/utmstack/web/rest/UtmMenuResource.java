package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmMenu;
import com.park.utmstack.domain.UtmMenuAuthority;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.shared_types.MenuType;
import com.park.utmstack.service.UtmMenuAuthorityService;
import com.park.utmstack.service.UtmMenuQueryService;
import com.park.utmstack.service.UtmMenuService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmMenuCriteria;
import com.park.utmstack.util.exceptions.UtmEntityRemoveException;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import io.micrometer.core.annotation.Timed;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api")
public class UtmMenuResource {

    private final Logger log = LoggerFactory.getLogger(UtmMenuResource.class);
    private static final String ENTITY_NAME = "UtmMenu";
    private static final String CLASS_NAME = "UtmMenuResource";

    private final UtmMenuQueryService menuQueryService;
    private final UtmMenuService menuService;
    private final UtmMenuAuthorityService menuAuthorityService;
    private final ApplicationEventService applicationEventService;

    public UtmMenuResource(UtmMenuQueryService menuQueryService, UtmMenuService menuService,
                           UtmMenuAuthorityService menuAuthorityService,
                           ApplicationEventService applicationEventService) {
        this.menuQueryService = menuQueryService;
        this.menuService = menuService;
        this.menuAuthorityService = menuAuthorityService;
        this.applicationEventService = applicationEventService;
    }

    @PostMapping("/menu")
    @Timed
    public ResponseEntity<UtmMenu> createUtmMenu(@Valid @RequestBody UtmMenu menu) {
        final String ctx = CLASS_NAME + ".createUtmMenu";

        try {
            if (menu.getId() != null)
                throw new BadRequestAlertException("A new UtmMenu cannot already have an ID", ENTITY_NAME, "idexists");

            UtmMenu result = menuService.save(menu);

            if (!CollectionUtils.isEmpty(menu.getAuthorities()))
                saveOrUpdateMenuAuthorities(result.getId(), menu.getAuthorities());

            return ResponseEntity.created(new URI("/api/menu/" + result.getId())).headers(
                HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString())).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }

    }

    @PutMapping("/menu")
    @Timed
    public ResponseEntity<UtmMenu> updateUtmMenu(@Valid @RequestBody UtmMenu menu) {
        final String ctx = CLASS_NAME + ".updateUtmMenu";

        try {
            if (menu.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmMenu result = menuService.save(menu);
            menuAuthorityService.deleteByMenuId(result.getId());

            if (!CollectionUtils.isEmpty(menu.getAuthorities()))
                saveOrUpdateMenuAuthorities(result.getId(), menu.getAuthorities());

            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }


    /**
     * @param criteria
     * @return
     */
    @GetMapping("/menu")
    @Timed
    public ResponseEntity<List<UtmMenu>> getAllUtmMenu(UtmMenuCriteria criteria) {
        final String ctx = CLASS_NAME + ".getAllUtmMenu";
        try {
            PageRequest sortByPosition = PageRequest.of(0, 10000, Sort.by(Sort.Order.asc("position")));
            Page<UtmMenu> page = menuQueryService.findByCriteria(criteria, sortByPosition);

            List<UtmMenu> menus = new ArrayList<>();
            if (page != null && !CollectionUtils.isEmpty(page.getContent())) {
                menus = page.getContent();
                menus.forEach(menu -> menu.setAuthorities(buildAuthorities(menu.getId())));
            }
            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", String.valueOf(menus.size()));
            return new ResponseEntity<>(menus, headers, HttpStatus.OK);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * @param id
     * @return
     * @throws UtmEntityRemoveException
     */
    @DeleteMapping("/menu/{id}")
    @Timed
    public ResponseEntity<Void> deleteUtmMenu(@PathVariable Long id) throws UtmEntityRemoveException {
        String ctx = CLASS_NAME + ".deleteUtmMenus";
        try {
            menuService.delete(id);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * @param parentId
     * @param authorities
     * @return
     */
    @GetMapping("/menu/get-menu-by-authorities")
    @Timed
    public ResponseEntity<List<UtmMenu>> getMenusByAuthorities(@RequestParam Long parentId,
                                                               @RequestParam List<String> authorities) {
        String ctx = CLASS_NAME + ".getMenusByAuthorities";
        try {
            List<UtmMenu> menus = menuService.findMenusByAuthorities(parentId, authorities).orElse(new ArrayList<>());
            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", String.valueOf(menus.size()));

            return new ResponseEntity<>(menus, headers, HttpStatus.OK);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * @param menuId
     * @param authorities
     */
    private void saveOrUpdateMenuAuthorities(Long menuId, List<String> authorities) {
        String ctx = CLASS_NAME + ".saveOrUpdateMenuAuthorities";
        try {
            authorities.forEach(auth -> {
                UtmMenuAuthority authority = new UtmMenuAuthority();
                authority.setAuthorityName(auth);
                authority.setMenuId(menuId);
                menuAuthorityService.save(authority);
            });
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * @param menuId
     * @return
     */
    private List<String> buildAuthorities(Long menuId) {
        String ctx = CLASS_NAME + ".buildAuthorities";
        try {
            Optional<List<UtmMenuAuthority>> authorities = menuAuthorityService.findAllByMenuIdEquals(menuId);
            List<String> result = new ArrayList<>();
            authorities.ifPresent(utmMenuAuthorities -> {
                utmMenuAuthorities.forEach(utmMenuAuthority -> result.add(utmMenuAuthority.getAuthorityName()));
            });
            return result;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    @GetMapping("/menu/all")
    public ResponseEntity<List<MenuType>> getMenus(@RequestParam boolean includeModulesMenus) {
        String ctx = CLASS_NAME + ".getMenus";
        try {
            return ResponseEntity.ok(menuService.getMenus(includeModulesMenus));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/menu/save-menu-structure")
    public ResponseEntity<Boolean> saveMenuStructure(@RequestBody List<MenuType> menus) {
        String ctx = CLASS_NAME + ".saveMenuStructure";
        try {
            return ResponseEntity.ok(menuService.saveMenuStructure(menus));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}
