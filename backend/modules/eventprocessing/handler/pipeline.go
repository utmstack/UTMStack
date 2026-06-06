package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
)

type PipelineHandler struct {
	uc connectors.PipelineUsecase
}

func NewPipelineHandler(uc connectors.PipelineUsecase) *PipelineHandler {
	return &PipelineHandler{uc: uc}
}

// @Summary     List pipelines
// @Tags        Logstash Pipelines
// @Security    BearerAuth
// @Produce     json
// @Param       page query int false "Page (0-based)"
// @Param       size query int false "Page size"
// @Success     200 {array} dto.UtmLogstashPipelineDTO
// @Header      200 {string} X-Total-Count "Total number of items"
// @Failure     500 {object} map[string]string
// @Router      /logstash-pipelines [get]
func (h *PipelineHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	items, total, err := h.uc.List(c.Request.Context(), page, size)
	if err != nil {
		logHandlerError("listLogstashPipelines", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writePagedArray(c, items, total)
}

// @Summary     Get pipeline stats
// @Tags        Logstash Pipelines
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} dto.ApiStatsResponse
// @Failure     500 {object} map[string]string
// @Router      /logstash-pipelines/stats [get]
func (h *PipelineHandler) GetStats(c *gin.Context) {
	stats, err := h.uc.GetStats(c.Request.Context())
	if err != nil {
		logHandlerError("getLogstashPipelineStats", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// @Summary     Get pipeline by ID
// @Tags        Logstash Pipelines
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Pipeline ID"
// @Success     200 {object} dto.UtmLogstashPipelineVM
// @Failure     500 {object} map[string]string
// @Router      /logstash-pipelines/{id} [get]
func (h *PipelineHandler) GetByID(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	vm, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		logHandlerError("getLogstashPipelineById", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vm)
}

// @Summary     Validate pipeline
// @Tags        Logstash Pipelines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       mode  query string                  true  "Validation mode (INSERT or UPDATE)"
// @Param       input body  dto.UtmLogstashPipelineVM true "Pipeline view model"
// @Success     204 "Valid"
// @Failure     500 {object} dto.PipelineErrors
// @Router      /logstash-pipelines/validate [post]
func (h *PipelineHandler) Validate(c *gin.Context) {
	mode := c.Query("mode")

	var vm dto.UtmLogstashPipelineVM
	if err := c.ShouldBindJSON(&vm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	pipelineErrors, err := h.uc.Validate(c.Request.Context(), vm, mode)
	if err != nil {
		logHandlerError("validateLogstashPipeline", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pipelineErrors == nil || len(pipelineErrors.ValidationErrors) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusInternalServerError, pipelineErrors)
}

// @Summary     Delete pipeline
// @Tags        Logstash Pipelines
// @Security    BearerAuth
// @Param       id path int true "Pipeline ID"
// @Success     200 "Deleted"
// @Failure     500 {object} map[string]string
// @Router      /logstash-pipelines/{id} [delete]
func (h *PipelineHandler) Delete(c *gin.Context) {
	id, ok := pathInt64(c, "id")
	if !ok {
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		logHandlerError("deleteLogstashPipeline", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
