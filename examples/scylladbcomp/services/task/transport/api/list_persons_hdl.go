package api

import (
	"net/http"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

// ScyllaListPersonHdl
func (a *api) ScyllaListPersonHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("auth-service").Start(c.Request.Context(), "ScyllaListPersonHdl")
		defer span.End()

		var filter entity.PersonFilter

		if err := c.ShouldBind(&filter); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		persons, err := a.biz.ScyllaListPersons(ctx, &filter)

		if err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.SuccessResponse(persons, nil, filter))
	}
}
