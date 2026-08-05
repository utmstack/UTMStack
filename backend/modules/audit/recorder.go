package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
)

var auditLogger connectors.Logger

func initRecorder(l connectors.Logger) {
	if auditLogger == nil {
		auditLogger = l
	}
}

func Record(c *gin.Context, ev connectors.Event, attemptType, successType domain.ApplicationEventType, err error) {
	if auditLogger == nil {
		return
	}
	if err != nil {
		ev.EventType = attemptType
		ev.Status = domain.StatusFailure
		ev.ErrorMessage = err.Error()
	} else {
		ev.EventType = successType
		ev.Status = domain.StatusSuccess
	}
	ev.IP = c.ClientIP()
	ev.UserAgent = c.Request.UserAgent()
	ev.UserLogin = c.GetString("user_login")
	if uid := c.GetUint64("user_id"); uid != 0 {
		ev.UserID = &uid
	}
	if sid := c.GetUint64("session_id"); sid != 0 {
		ev.SessionID = &sid
	}
	ev.SupportAccess = c.GetString("support_access")
	auditLogger.Log(c.Request.Context(), ev)
}
