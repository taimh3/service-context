package api

import (
	"net/http"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

// ScyllaCreateTaskHdl
func (a *api) ScyllaCreateTaskHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("auth-service").Start(c.Request.Context(), "ScyllaCreateTaskHdl")
		defer span.End()

		var data entity.TaskCreateRequest

		if err := c.ShouldBind(&data); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := a.biz.ScyllaAddNewTask(ctx, &data); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		data.Mask()
		c.JSON(http.StatusOK, core.ResponseData(data.FakeId))
	}
}
