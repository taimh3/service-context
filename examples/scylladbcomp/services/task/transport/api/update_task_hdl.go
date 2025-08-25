package api

import (
	"net/http"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

// ScyllaUpdateTaskHdl
func (a *api) ScyllaUpdateTaskHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("auth-service").Start(c.Request.Context(), "ScyllaUpdateTaskHdl")
		defer span.End()

		uid, err := core.FromBase58(c.Param("task-id"))

		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		var data entity.TaskUpdateRequest

		if err := c.ShouldBind(&data); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := a.biz.ScyllaUpdateTask(ctx, int(uid.GetLocalID()), &data); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData(true))
	}
}
