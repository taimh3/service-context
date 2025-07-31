package core

// Error Response Structures
type ErrorDetail struct {
	Code        string `json:"code"`
	Title       string `json:"title,omitempty"` // optional title for the error
	Message     string `json:"message"`
	Description string `json:"description,omitempty"` // debug only
}

type StandardResponse struct {
	Code  string       `json:"code,omitempty"` // action directive, ex: REDIRECT_HOME, POPUP_MESSAGE
	Data  interface{}  `json:"data,omitempty"` // data response (array or object)
	Error *ErrorDetail `json:"error,omitempty"`
}

// DebugContext interface for handling debug mode
type DebugContext interface {
	IsDebugEnabled() bool
}

// GlobalDebugContext holds the global debug configuration
var GlobalDebugContext DebugContext

// SetGlobalDebugContext sets the global debug context
func SetGlobalDebugContext(ctx DebugContext) {
	GlobalDebugContext = ctx
}

// isDebugMode returns true if debug mode is enabled
func isDebugMode() bool {
	if GlobalDebugContext != nil {
		return GlobalDebugContext.IsDebugEnabled()
	}
	return false
}

// ErrorContext provides context-aware error response functions
type ErrorContext struct {
	debugEnabled bool
}

// Singleton instance for global error context
var GlobalErrorContext *ErrorContext

// InitGlobalErrorContext initializes the global error context with auto debug detection
func InitGlobalErrorContext() {
	if GlobalErrorContext != nil {
		return // already initialized
	}
	GlobalErrorContext = newErrorContextAuto()
}

// newErrorContextAuto creates a new error context with auto debug detection
func newErrorContextAuto() *ErrorContext {
	return &ErrorContext{debugEnabled: isDebugMode()}
}

// NewErrorContext creates a new ErrorContext with BadRequest error details
func (ec *ErrorContext) BadRequestError(message, description string) StandardResponse {
	return NewErrorResponse("BAD_REQUEST", "", message, description, ec.debugEnabled)
}

// NewErrorResponse creates a new StandardResponse with Unauthorized error details
func (ec *ErrorContext) UnauthorizedError(message, description string) StandardResponse {
	return NewErrorResponse("UNAUTHORIZED", "", message, description, ec.debugEnabled)
}

// NewErrorResponse create a new StandardResponse with forbidden error details
func (ec *ErrorContext) ForbiddenError(message, description string) StandardResponse {
	return NewErrorResponse("FORBIDDEN", "", message, description, ec.debugEnabled)
}

// NewErrorResponse creates a new StandardResponse with not found error details
func (ec *ErrorContext) NotFoundError(message, description string) StandardResponse {
	return NewErrorResponse("NOT_FOUND", "", message, description, ec.debugEnabled)
}

// NewErrorResponse creates a new StandardResponse with Internal Server Error details
func (ec *ErrorContext) InternalServerError(message, description string) StandardResponse {
	return NewErrorResponse("INTERNAL_ERROR", "", message, description, ec.debugEnabled)
}

// NewErrorResponse creates a new StandardResponse with error details
func (ec *ErrorContext) ServiceUnavailableError(message, description string) StandardResponse {
	return NewErrorResponse("SERVICE_UNAVAILABLE", "", message, description, ec.debugEnabled)
}

// ServiceNotImplementedError returns a standard response for service not implemented errors
func (ec *ErrorContext) ServiceNotImplementedError(message, description string) StandardResponse {
	return NewErrorResponse("SERVICE_NOT_IMPLEMENTED", "", message, description, ec.debugEnabled)
}

// NewErrorResponse creates a new StandardResponse with custom error details
func (ec *ErrorContext) CustomError(code, title, message, description string, actionCode ...string) StandardResponse {
	return NewErrorResponse(code, title, message, description, ec.debugEnabled, actionCode...)
}
