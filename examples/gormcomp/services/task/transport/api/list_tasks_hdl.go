package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
)

// ListTaskHdl handles the request to list tasks
func (a *api) ListTaskHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		tasks, err := a.biz.ListTasks(c.Request.Context(), nil, nil)
		if err != nil {
			core.WriteStandardErrorResponse(c, http.StatusInternalServerError, core.GlobalErrorContext.InternalServerError(
				core.ErrInternalServerError.Error(),
				err.Error(),
			))
			return
		}

		core.WriteSuccessResponse(c, tasks)
	}
}
