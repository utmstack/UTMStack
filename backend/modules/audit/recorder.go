package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
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
	// Only when the caller did not name someone themselves: on the login route
	// there is no actor yet, and the address typed is the only identity there is.
	if ev.UserEmail == "" {
		ev.UserEmail = c.GetString("user_email")
	}
	if a := middleware.ActorFromGin(c); a != nil {
		if a.UserID != uuid.Nil {
			id := a.UserID
			ev.UserID = &id
		}
		if a.SessionID != uuid.Nil {
			sid := a.SessionID
			ev.SessionID = &sid
		}
	}
	ev.SupportAccess = c.GetString("support_access")
	auditLogger.Log(c.Request.Context(), ev)
}
