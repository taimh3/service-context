package api

import (
	"net/http"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

// ScyllaListTaskHdl
func (a *api) ScyllaListTaskHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("auth-service").Start(c.Request.Context(), "ScyllaListTaskHdl")
		defer span.End()

		type reqParam struct {
			entity.Filter
			core.Paging
		}

		var rp reqParam

		if err := c.ShouldBind(&rp); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		rp.Paging.Process()

		tasks, err := a.biz.ScyllaListTasks(ctx, &rp.Filter, &rp.Paging)

		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		for i := range tasks {
			tasks[i].Mask()
		}

		c.JSON(http.StatusOK, core.SuccessResponse(tasks, rp.Paging, rp.Filter))
	}
}
