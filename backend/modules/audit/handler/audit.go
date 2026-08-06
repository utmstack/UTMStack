package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/dto"
)

type Handler struct {
	usecase connectors.Usecase
}

func NewHandler(uc connectors.Usecase) *Handler {
	return &Handler{usecase: uc}
}

// @Summary     List audit log entries
// @Tags        Audit
// @Security    BearerAuth
// @Produce     json
// @Param       search_query   query string false "filter DSL: field.op.value joined by & (e.g. status.eq.success&timestamp.gte.2026-06-01). Allowed fields: user_id, user_email, action, status, resource_type, resource_id, event_type, ip, timestamp. Ops: eq,neq,like,nlike,in,nin,gt,gte,lt,lte,null,nnull"
// @Param       sort_by        query string false "sort DSL: field.dir joined by & (e.g. timestamp.desc). Same allowed fields. Default: id desc"
// @Param       page_number    query int    false "default 1"
// @Param       page_size      query int    false "default 50, max 500"
// @Success     200 {object} dto.ListResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /audit [get]
func (h *Handler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.usecase.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Get audit log entry by id
// @Tags        Audit
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Audit log id"
// @Success     200 {object} dto.LogResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /audit/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.usecase.Get(c.Request.Context(), id)
	if err != nil {
		_ = catcher.Error("audit get failed", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get audit log"})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
