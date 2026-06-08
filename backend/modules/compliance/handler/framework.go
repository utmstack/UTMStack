package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit"
	audit_connectors "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type FrameworkHandler struct{ uc connectors.FrameworkUsecase }

func NewFrameworkHandler(uc connectors.FrameworkUsecase) *FrameworkHandler {
	return &FrameworkHandler{uc: uc}
}

// ── Control library (read) ────────────────────────────────────────────────────

// ListControls godoc
//
//	@Summary     List compliance controls
//	@Description Returns the full reusable control library (NIST 800-53 hub, system + user overlay).
//	@Tags        Compliance Controls
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {array}  domain.Control
//	@Failure     500 {object} map[string]string
//	@Router      /compliance/controls [get]
func (h *FrameworkHandler) ListControls(c *gin.Context) {
	c.JSON(http.StatusOK, h.uc.ListControls(c.Request.Context()))
}

// GetControl godoc
//
//	@Summary     Get a compliance control
//	@Description Returns a single control by id (e.g. "au-6").
//	@Tags        Compliance Controls
//	@Security    BearerAuth
//	@Produce     json
//	@Param       id  path     string true "Control id"
//	@Success     200 {object} domain.Control
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/controls/{id} [get]
func (h *FrameworkHandler) GetControl(c *gin.Context) {
	item, err := h.uc.GetControl(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// ── Control library (user-overlay CRUD) ───────────────────────────────────────

// CreateControl godoc
//
//	@Summary     Create a custom control
//	@Description Creates a user control in the user overlay. Fails (409) if the id already exists.
//	@Tags        Compliance Controls
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body     domain.Control true "Control to create"
//	@Success     201  {object} domain.Control
//	@Failure     400  {object} map[string]string
//	@Failure     409  {object} map[string]string
//	@Router      /compliance/controls [post]
func (h *FrameworkHandler) CreateControl(c *gin.Context) {
	var ctrl domain.Control
	if err := c.ShouldBindJSON(&ctrl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.uc.CreateControl(c.Request.Context(), ctrl)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.create", ResourceType: "compliance_control", ResourceID: ctrl.ID},
		audit_domain.COMPLIANCE_CONTROL_CREATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// UpdateControl godoc
//
//	@Summary     Update a control (copy-on-write)
//	@Description Writes the control to the user overlay. Editing a vendor (system) control creates a user override; the system file is untouched.
//	@Tags        Compliance Controls
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       id   path     string        true "Control id"
//	@Param       body body     domain.Control true "Control body"
//	@Success     200  {object} domain.Control
//	@Failure     400  {object} map[string]string
//	@Failure     404  {object} map[string]string
//	@Router      /compliance/controls/{id} [put]
func (h *FrameworkHandler) UpdateControl(c *gin.Context) {
	var ctrl domain.Control
	if err := c.ShouldBindJSON(&ctrl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctrl.ID = c.Param("id") // path id is authoritative
	out, err := h.uc.UpdateControl(c.Request.Context(), ctrl)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.update", ResourceType: "compliance_control", ResourceID: ctrl.ID},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// DeleteControl godoc
//
//	@Summary     Delete a custom control
//	@Description Removes the user-overlay file. A vendor-only control cannot be deleted (403) — disable it instead.
//	@Tags        Compliance Controls
//	@Security    BearerAuth
//	@Param       id  path string true "Control id"
//	@Success     204
//	@Failure     403 {object} map[string]string
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/controls/{id} [delete]
func (h *FrameworkHandler) DeleteControl(c *gin.Context) {
	id := c.Param("id")
	err := h.uc.DeleteControl(c.Request.Context(), id)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.delete", ResourceType: "compliance_control", ResourceID: id},
		audit_domain.COMPLIANCE_CONTROL_DELETE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetControlEnabled godoc
//
//	@Summary     Enable/disable a control
//	@Description Toggles the control's enabled state via the .disabled filename suffix.
//	@Tags        Compliance Controls
//	@Security    BearerAuth
//	@Accept      json
//	@Param       id   path string             true "Control id"
//	@Param       body body dto.EnabledRequest true "Enabled flag"
//	@Success     204
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/controls/{id}/enabled [put]
func (h *FrameworkHandler) SetControlEnabled(c *gin.Context) {
	id := c.Param("id")
	var req dto.EnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.uc.SetControlEnabled(c.Request.Context(), id, req.Enabled)
	audit.Record(c, audit_connectors.Event{Action: "compliance.control.set-enabled", ResourceType: "compliance_control", ResourceID: id},
		audit_domain.COMPLIANCE_CONTROL_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_CONTROL_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Frameworks (read) ─────────────────────────────────────────────────────────

// ListFrameworks godoc
//
//	@Summary     List compliance frameworks
//	@Description Returns the framework checklists (HIPAA, PCI, CMMC, …) that reference controls.
//	@Tags        Compliance Frameworks
//	@Security    BearerAuth
//	@Produce     json
//	@Success     200 {array}  domain.Framework
//	@Failure     500 {object} map[string]string
//	@Router      /compliance/frameworks [get]
func (h *FrameworkHandler) ListFrameworks(c *gin.Context) {
	c.JSON(http.StatusOK, h.uc.ListFrameworks(c.Request.Context()))
}

// GetFramework godoc
//
//	@Summary     Get a compliance framework
//	@Description Returns a single framework by key (e.g. "hipaa").
//	@Tags        Compliance Frameworks
//	@Security    BearerAuth
//	@Produce     json
//	@Param       key path     string true "Framework key"
//	@Success     200 {object} domain.Framework
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key} [get]
func (h *FrameworkHandler) GetFramework(c *gin.Context) {
	fw, err := h.uc.GetFramework(c.Request.Context(), c.Param("key"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, fw)
}

// ── Frameworks (user-overlay CRUD) ────────────────────────────────────────────

// CreateFramework godoc
//
//	@Summary     Create a custom framework
//	@Description Creates a user framework in the user overlay. Fails (409) if the key already exists.
//	@Tags        Compliance Frameworks
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       body body     domain.Framework true "Framework to create"
//	@Success     201  {object} domain.Framework
//	@Failure     400  {object} map[string]string
//	@Failure     409  {object} map[string]string
//	@Router      /compliance/frameworks [post]
func (h *FrameworkHandler) CreateFramework(c *gin.Context) {
	var fw domain.Framework
	if err := c.ShouldBindJSON(&fw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.uc.CreateFramework(c.Request.Context(), fw)
	audit.Record(c, audit_connectors.Event{Action: "compliance.framework.create", ResourceType: "compliance_framework", ResourceID: fw.Key},
		audit_domain.COMPLIANCE_FRAMEWORK_CREATE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_CREATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// UpdateFramework godoc
//
//	@Summary     Update a framework (copy-on-write)
//	@Description Writes the framework to the user overlay. Editing a vendor framework creates a user override.
//	@Tags        Compliance Frameworks
//	@Security    BearerAuth
//	@Accept      json
//	@Produce     json
//	@Param       key  path     string           true "Framework key"
//	@Param       body body     domain.Framework true "Framework body"
//	@Success     200  {object} domain.Framework
//	@Failure     400  {object} map[string]string
//	@Failure     404  {object} map[string]string
//	@Router      /compliance/frameworks/{key} [put]
func (h *FrameworkHandler) UpdateFramework(c *gin.Context) {
	var fw domain.Framework
	if err := c.ShouldBindJSON(&fw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fw.Key = c.Param("key")
	out, err := h.uc.UpdateFramework(c.Request.Context(), fw)
	audit.Record(c, audit_connectors.Event{Action: "compliance.framework.update", ResourceType: "compliance_framework", ResourceID: fw.Key},
		audit_domain.COMPLIANCE_FRAMEWORK_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// DeleteFramework godoc
//
//	@Summary     Delete a custom framework
//	@Description Removes the user-overlay file. A vendor-only framework cannot be deleted (403) — disable it instead.
//	@Tags        Compliance Frameworks
//	@Security    BearerAuth
//	@Param       key path string true "Framework key"
//	@Success     204
//	@Failure     403 {object} map[string]string
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key} [delete]
func (h *FrameworkHandler) DeleteFramework(c *gin.Context) {
	key := c.Param("key")
	err := h.uc.DeleteFramework(c.Request.Context(), key)
	audit.Record(c, audit_connectors.Event{Action: "compliance.framework.delete", ResourceType: "compliance_framework", ResourceID: key},
		audit_domain.COMPLIANCE_FRAMEWORK_DELETE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_DELETE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetFrameworkEnabled godoc
//
//	@Summary     Enable/disable a framework
//	@Description Toggles the framework's enabled state via the .disabled filename suffix.
//	@Tags        Compliance Frameworks
//	@Security    BearerAuth
//	@Accept      json
//	@Param       key  path string             true "Framework key"
//	@Param       body body dto.EnabledRequest true "Enabled flag"
//	@Success     204
//	@Failure     404 {object} map[string]string
//	@Router      /compliance/frameworks/{key}/enabled [put]
func (h *FrameworkHandler) SetFrameworkEnabled(c *gin.Context) {
	key := c.Param("key")
	var req dto.EnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.uc.SetFrameworkEnabled(c.Request.Context(), key, req.Enabled)
	audit.Record(c, audit_connectors.Event{Action: "compliance.framework.set-enabled", ResourceType: "compliance_framework", ResourceID: key},
		audit_domain.COMPLIANCE_FRAMEWORK_UPDATE_ATTEMPT, audit_domain.COMPLIANCE_FRAMEWORK_UPDATE_SUCCESS, err)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
