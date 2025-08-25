package api

import (
	"net/http"

	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/gin-gonic/gin"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

// ScyllaCreatePersonHdl
func (a *api) ScyllaCreatePersonHdl() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("auth-service").Start(c.Request.Context(), "ScyllaCreatePersonHdl")
		defer span.End()

		var data entity.PersonCreateRequest

		if err := c.ShouldBind(&data); err != nil {
			common.WriteErrorResponse(c, core.ErrBadRequest.WithError(err.Error()))
			return
		}

		if err := a.biz.ScyllaAddNewPerson(ctx, &data); err != nil {
			common.WriteErrorResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, core.ResponseData("Person created successfully"))
	}
}
