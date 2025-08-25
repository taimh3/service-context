package api

import (
	"net/http"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

// ScyllaDeleteTaskHdl
func (a *api) ScyllaDeleteTaskHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("auth-service").Start(c.Request.Context(), "ScyllaDeleteTaskHdl")
		defer span.End()

		uid, err := core.FromBase58(c.Param("task-id"))

		if err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := a.biz.ScyllaDeleteTask(ctx, int(uid.GetLocalID())); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData(true))
	}
}
