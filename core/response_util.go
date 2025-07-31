package core

import "github.com/gin-gonic/gin"

// NewSuccessResponse creates a new success response with optional action code
func NewSuccessResponse(data interface{}, actionCode ...string) StandardResponse {
	response := StandardResponse{
		Data: data,
	}

	if len(actionCode) > 0 && actionCode[0] != "" {
		response.Code = actionCode[0]
	}

	return response
}

// NewErrorResponse creates a new error response with optional action code
func NewErrorResponse(code, title, message, description string, showDebug bool, actionCode ...string) StandardResponse {
	desc := ""
	if showDebug {
		desc = description
	}

	response := StandardResponse{
		Error: &ErrorDetail{
			Code:        code,
			Title:       title,
			Message:     message,
			Description: desc,
		},
	}

	if len(actionCode) > 0 && actionCode[0] != "" {
		response.Code = actionCode[0]
	}

	return response
}

// WriteStandardErrorResponse writes error using the new standard format
func WriteStandardErrorResponse(c *gin.Context, httpStatus int, errResponse StandardResponse) {
	c.JSON(httpStatus, errResponse)
}
