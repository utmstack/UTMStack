package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/edr/domain"
	"github.com/utmstack/utmstack/backend/modules/edr/dto"
	"github.com/utmstack/utmstack/backend/modules/edr/repository"
	"github.com/utmstack/utmstack/backend/modules/edr/usecase"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

type Handler struct {
	svc *usecase.Service
}

func NewHandler(svc *usecase.Service) *Handler {
	return &Handler{svc: svc}
}

// @Summary EDR fleet overview
// @Tags EDR
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.OverviewResponse
// @Router /edr/overview [get]
func (h *Handler) Overview(c *gin.Context) {
	out, err := h.svc.Overview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// @Summary List EDR endpoints
// @Tags EDR
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.Endpoint
// @Router /edr/endpoints [get]
func (h *Handler) Endpoints(c *gin.Context) {
	out, err := h.svc.Endpoints(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// @Summary One EDR endpoint
// @Tags EDR
// @Security BearerAuth
// @Param agentId path int true "agent-manager id"
// @Produce json
// @Success 200 {object} dto.Endpoint
// @Failure 404 {object} map[string]string
// @Router /edr/endpoints/{agentId} [get]
func (h *Handler) Endpoint(c *gin.Context) {
	id, err := parseAgentID(c.Param("agentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agentId"})
		return
	}
	out, err := h.svc.Endpoint(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrEndpointNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

type actionRequest struct {
	Action  string          `json:"action" binding:"required"`
	Payload json.RawMessage `json:"payload"`
	Reason  string          `json:"reason"`
}

// @Summary Dispatch an EDR action to one endpoint
// @Tags EDR
// @Security BearerAuth
// @Param agentId path int true "agent-manager id"
// @Param body body actionRequest true "action + optional JSON payload"
// @Produce json
// @Success 200 {object} dto.ActionResult
// @Router /edr/endpoints/{agentId}/actions [post]
func (h *Handler) Action(c *gin.Context) {
	id, err := parseAgentID(c.Param("agentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agentId"})
		return
	}
	var body actionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	out, err := h.svc.Action(c.Request.Context(), id, body.Action, string(body.Payload), loginFromCtx(c), body.Reason)
	if err != nil {
		if out != nil {
			out.Error = err.Error()
			c.JSON(http.StatusBadGateway, out)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func parseAgentID(s string) (uint32, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// @Summary List EDR policies
// @Tags EDR
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.PolicyResponse
// @Router /edr/policies [get]
func (h *Handler) ListPolicies(c *gin.Context) {
	ps, err := h.svc.ListPolicies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]dto.PolicyResponse, 0, len(ps))
	for _, p := range ps {
		out = append(out, policyResp(&p))
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Create an EDR policy
// @Tags EDR
// @Security BearerAuth
// @Param body body dto.PolicyRequest true "policy document"
// @Produce json
// @Success 201 {object} dto.PolicyResponse
// @Router /edr/policies [post]
func (h *Handler) CreatePolicy(c *gin.Context) {
	var body dto.PolicyRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	p, err := h.svc.CreatePolicy(c.Request.Context(), body.Name, body.Description, body.Version, body.Doc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, policyResp(p))
}

// @Summary Get one EDR policy
// @Tags EDR
// @Security BearerAuth
// @Param id path string true "policy id"
// @Produce json
// @Success 200 {object} dto.PolicyResponse
// @Router /edr/policies/{id} [get]
func (h *Handler) GetPolicy(c *gin.Context) {
	id, err := parsePolicyID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	p, err := h.svc.GetPolicy(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPolicyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, policyResp(p))
}

// @Summary Update one EDR policy
// @Tags EDR
// @Security BearerAuth
// @Param id path string true "policy id"
// @Param body body dto.PolicyRequest true "policy document"
// @Produce json
// @Success 200 {object} dto.PolicyResponse
// @Router /edr/policies/{id} [put]
func (h *Handler) UpdatePolicy(c *gin.Context) {
	id, err := parsePolicyID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body dto.PolicyRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	p, err := h.svc.UpdatePolicy(c.Request.Context(), id, body.Name, body.Description, body.Version, body.Doc)
	if err != nil {
		if errors.Is(err, repository.ErrPolicyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, policyResp(p))
}

// @Summary Delete one EDR policy
// @Tags EDR
// @Security BearerAuth
// @Param id path string true "policy id"
// @Produce json
// @Success 204
// @Router /edr/policies/{id} [delete]
func (h *Handler) DeletePolicy(c *gin.Context) {
	id, err := parsePolicyID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeletePolicy(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type assignRequest struct {
	AgentIDs []uint32 `json:"agentIds" binding:"required"`
	Reason   string   `json:"reason"`
}

// @Summary Assign a policy to endpoints
// @Tags EDR
// @Security BearerAuth
// @Param id path string true "policy id"
// @Param body body assignRequest true "target endpoints"
// @Produce json
// @Success 200 {array} dto.ActionResult
// @Router /edr/policies/{id}/assignments [put]
func (h *Handler) AssignPolicy(c *gin.Context) {
	id, err := parsePolicyID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body assignRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	out, err := h.svc.AssignPolicy(c.Request.Context(), id, body.AgentIDs, loginFromCtx(c), body.Reason)
	if err != nil {
		if errors.Is(err, repository.ErrPolicyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// @Summary List quarantined items
// @Tags EDR
// @Security BearerAuth
// @Param agentId query int false "filter to one endpoint"
// @Produce json
// @Success 200 {array} dto.QuarantineItem
// @Router /edr/quarantine [get]
func (h *Handler) Quarantine(c *gin.Context) {
	var filter *uint32
	if q := c.Query("agentId"); q != "" {
		if id, err := parseAgentID(q); err == nil {
			filter = &id
		}
	}
	out, err := h.svc.Quarantine(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Restore a quarantined item
// @Tags EDR
// @Security BearerAuth
// @Param itemId path string true "quarantine id"
// @Param agentId query int true "endpoint that holds the item"
// @Produce json
// @Success 200 {object} dto.ActionResult
// @Router /edr/quarantine/{itemId}/restore [post]
func (h *Handler) QuarantineRestore(c *gin.Context) { h.quarantineAct("restore")(c) }

// @Summary Purge a quarantined item
// @Tags EDR
// @Security BearerAuth
// @Param itemId path string true "quarantine id"
// @Param agentId query int true "endpoint that holds the item"
// @Produce json
// @Success 200 {object} dto.ActionResult
// @Router /edr/quarantine/{itemId}/purge [post]
func (h *Handler) QuarantinePurge(c *gin.Context) { h.quarantineAct("purge")(c) }

func (h *Handler) quarantineAct(op string) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentIDStr := c.Query("agentId")
		if agentIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "agentId query is required"})
			return
		}
		id, err := parseAgentID(agentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agentId"})
			return
		}
		out, err := h.svc.QuarantineItem(c.Request.Context(), id, c.Param("itemId"), op, loginFromCtx(c))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	}
}

// @Summary Blocklist feed health
// @Tags EDR
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.BlocklistStatus
// @Router /edr/blocklist/status [get]
func (h *Handler) BlocklistStatus(c *gin.Context) {
	out, err := h.svc.BlocklistStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Allowlist entries
// @Tags EDR
// @Security BearerAuth
// @Param agentId query int false "filter to one endpoint"
// @Produce json
// @Success 200 {array} dto.AllowlistEntry
// @Router /edr/blocklist/allowlist [get]
func (h *Handler) Allowlist(c *gin.Context) {
	var filter *uint32
	if q := c.Query("agentId"); q != "" {
		if id, err := parseAgentID(q); err == nil {
			filter = &id
		}
	}
	out, err := h.svc.Allowlist(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

type allowBody struct {
	Category string   `json:"category" binding:"required"`
	Value    string   `json:"value" binding:"required"`
	AgentIDs []uint32 `json:"agentIds" binding:"required"`
	Reason   string   `json:"reason"`
}

// pushAllow dispatches allow_add/allow_remove for one value to a set of
// endpoints. The agent's allow_* action takes {category,value} in its payload.
func (h *Handler) pushAllow(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body allowBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		payload, _ := json.Marshal(map[string]string{"category": body.Category, "value": body.Value})
		results := make([]dto.ActionResult, 0, len(body.AgentIDs))
		for _, id := range body.AgentIDs {
			res := dto.ActionResult{AgentID: id, Action: action}
			out, err := h.svc.Action(c.Request.Context(), id, action, string(payload), loginFromCtx(c), body.Reason)
			if err != nil {
				res.Error = err.Error()
			} else {
				res.CmdID = out.CmdID
				res.Dispatch = out.Dispatch
			}
			results = append(results, res)
		}
		c.JSON(http.StatusOK, results)
	}
}

// @Summary Add an allowlist entry to endpoints
// @Tags EDR
// @Security BearerAuth
// @Param body body allowBody true "category + value + targets"
// @Produce json
// @Success 200 {array} dto.ActionResult
// @Router /edr/blocklist/allowlist [post]
func (h *Handler) AllowAdd(c *gin.Context) { h.pushAllow("allow_add")(c) }

// @Summary Remove an allowlist entry from endpoints
// @Tags EDR
// @Security BearerAuth
// @Param body body allowBody true "category + value + targets"
// @Produce json
// @Success 200 {array} dto.ActionResult
// @Router /edr/blocklist/allowlist [delete]
func (h *Handler) AllowRemove(c *gin.Context) { h.pushAllow("allow_remove")(c) }

func policyResp(p *domain.Policy) dto.PolicyResponse {
	return dto.PolicyResponse{
		ID:          strconv.FormatUint(uint64(p.ID), 10),
		Name:        p.Name,
		Description: p.Description,
		Version:     p.Version,
		Doc:         p.Doc,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func loginFromCtx(c *gin.Context) string {
	if a := middleware.ActorFromGin(c); a != nil {
		return a.Email
	}
	return "edr-console"
}

func parsePolicyID(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	return uint(n), err
}
