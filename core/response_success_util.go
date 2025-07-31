package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WriteSuccessResponse writes a successful response with the given data using standard format
func WriteSuccessResponse(c *gin.Context, data interface{}, actionCode ...string) {
	c.JSON(http.StatusOK, NewSuccessResponse(data, actionCode...))
}

// WriteCreateSuccessResponse writes a successful creation response with the given data using standard format
func WriteCreateSuccessResponse(c *gin.Context, data interface{}, actionCode ...string) {
	c.JSON(http.StatusCreated, NewSuccessResponse(data, actionCode...))
}

// WriteNoContentSuccessResponse writes a successful no content response
func WriteNoContentSuccessResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
