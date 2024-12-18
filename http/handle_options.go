package http

import (
	"net/http"

	"github.com/my-ermes-labs/api-go/api"
)

// Options for the handler.
type HandlerOptions struct {
	getAcquireSessionOptions           func(req *http.Request) api.AcquireSessionOptions
	getCreateSessionOptions            func(req *http.Request) api.CreateSessionOptions
	getSessionTokenBytes               func(req *http.Request) []byte
	redirectNewRequest                 func(req *http.Request, node *api.Node) bool
	redirectTarget                     func(req *http.Request, node *api.Node) string
	setSessionTokenBytes               func(w http.ResponseWriter, sessionTokenBytes []byte)
	redirectResponse                   func(w http.ResponseWriter, req *http.Request, host string)
	malformedSessionTokenErrorResponse func(w http.ResponseWriter, err error)
	internalServerErrorResponse        func(w http.ResponseWriter, err error)
}

// Builder for HandlerOptions.
type HandlerOptionsBuilder struct {
	options HandlerOptions
}

// Create a new HandlerOptionsBuilder.
func NewHandlerOptionsBuilder() *HandlerOptionsBuilder {
	return &HandlerOptionsBuilder{
		options: DefaultHandlerOptions(),
	}
}

// Set the GetAcquireSessionOptions function.
func (builder *HandlerOptionsBuilder) GetAcquireSessionOptionsFunc(GetAcquireSessionOptions func(req *http.Request) api.AcquireSessionOptions) *HandlerOptionsBuilder {
	builder.options.getAcquireSessionOptions = GetAcquireSessionOptions
	return builder
}

// Set the AcquireSessionOptions function.
func (builder *HandlerOptionsBuilder) AcquireSessionOptions(acquireSessionOptions api.AcquireSessionOptions) *HandlerOptionsBuilder {
	builder.options.getAcquireSessionOptions = func(req *http.Request) api.AcquireSessionOptions {
		return acquireSessionOptions
	}
	return builder
}

// Set the CreateSessionOptions function.
func (builder *HandlerOptionsBuilder) GetCreateSessionOptionsFunc(CreateSessionOptions func(req *http.Request) api.CreateSessionOptions) *HandlerOptionsBuilder {
	builder.options.getCreateSessionOptions = CreateSessionOptions
	return builder
}

// Set the CreateSessionOptions function.
func (builder *HandlerOptionsBuilder) CreateSessionOptions(createSessionOptions api.CreateSessionOptions) *HandlerOptionsBuilder {
	builder.options.getCreateSessionOptions = func(req *http.Request) api.CreateSessionOptions {
		return createSessionOptions
	}
	return builder
}

// Set the getSessionTokenBytes function.
func (builder *HandlerOptionsBuilder) GetSessionTokenBytes(getSessionTokenBytes func(req *http.Request) []byte) *HandlerOptionsBuilder {
	builder.options.getSessionTokenBytes = getSessionTokenBytes
	return builder
}

// Set the redirectNewRequest function.
func (builder *HandlerOptionsBuilder) RedirectNewRequest(redirectNewRequest func(req *http.Request, node *api.Node) bool) *HandlerOptionsBuilder {
	builder.options.redirectNewRequest = redirectNewRequest
	return builder
}

// Set the redirectTarget function.
func (builder *HandlerOptionsBuilder) RedirectTarget(redirectTarget func(req *http.Request, node *api.Node) string) *HandlerOptionsBuilder {
	builder.options.redirectTarget = redirectTarget
	return builder
}

// Set the setSessionTokenBytes function.
func (builder *HandlerOptionsBuilder) SetSessionTokenBytes(setSessionTokenBytes func(w http.ResponseWriter, sessionTokenBytes []byte)) *HandlerOptionsBuilder {
	builder.options.setSessionTokenBytes = setSessionTokenBytes
	return builder
}

// Set the redirectResponse function.
func (builder *HandlerOptionsBuilder) RedirectResponse(redirectResponse func(w http.ResponseWriter, req *http.Request, host string)) *HandlerOptionsBuilder {
	builder.options.redirectResponse = redirectResponse
	return builder
}

// Set the malformedSessionTokenErrorResponse function.
func (builder *HandlerOptionsBuilder) MalformedSessionTokenErrorResponse(malformedSessionTokenErrorResponse func(w http.ResponseWriter, err error)) *HandlerOptionsBuilder {
	builder.options.malformedSessionTokenErrorResponse = malformedSessionTokenErrorResponse
	return builder
}

// Set the internalServerErrorResponse function.
func (builder *HandlerOptionsBuilder) InternalServerErrorResponse(internalServerErrorResponse func(w http.ResponseWriter, err error)) *HandlerOptionsBuilder {
	builder.options.internalServerErrorResponse = internalServerErrorResponse
	return builder
}

// Set the getSessionTokenBytes and setSessionTokenBytes functions to use the
// given header name to get and set the session token.
func (builder *HandlerOptionsBuilder) SessionTokenHeaderName(header string) *HandlerOptionsBuilder {
	builder.options.getSessionTokenBytes = func(req *http.Request) []byte {
		return GetSessionTokenBytesFromHeader(req, header)
	}
	builder.options.setSessionTokenBytes = func(w http.ResponseWriter, sessionTokenBytes []byte) {
		SetSessionTokenBytesToHeader(w, sessionTokenBytes, header)
	}
	return builder
}

// Build the HandlerOptions.
func (builder *HandlerOptionsBuilder) Build() HandlerOptions {
	return builder.options
}

// DefaultHandlerOptions returns the default options for the handler.
func DefaultHandlerOptions() HandlerOptions {
	return HandlerOptions{
		getAcquireSessionOptions: func(_ *http.Request) api.AcquireSessionOptions {
			// Return the default options to acquire a session.
			return api.DefaultAcquireSessionOptions()
		},
		getCreateSessionOptions: func(_ *http.Request) api.CreateSessionOptions {
			// Return the default options to create a session.
			return api.DefaultCreateSessionOptions()
		},
		getSessionTokenBytes: func(req *http.Request) []byte {
			return GetSessionTokenBytesFromHeader(req, DefaultTokenHeaderName)
		},
		redirectNewRequest: func(_ *http.Request, _ *api.Node) bool {
			// By default, do not redirect new requests.
			return false
		},
		redirectTarget: func(req *http.Request, node *api.Node) string {
			// By default, redirect to the node host.
			parent, err := node.GetParentNodeOf(req.Context(), node.Host)

			if err != nil {
				panic(err)
			}

			return parent.Host
		},
		setSessionTokenBytes: func(w http.ResponseWriter, sessionTokenBytes []byte) {
			SetSessionTokenBytesToHeader(w, sessionTokenBytes, DefaultTokenHeaderName)
		},
		redirectResponse: func(w http.ResponseWriter, req *http.Request, host string) {
			// Redirect the request to the given host.
			http.Redirect(w, req, host, http.StatusFound)
		},
		malformedSessionTokenErrorResponse: func(w http.ResponseWriter, err error) {
			// Return a bad request response with the error message.
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		internalServerErrorResponse: func(w http.ResponseWriter, err error) {
			// Return an internal server error response with the error message.
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}
}

// DefaultTokenHeaderName is the default name of the header that contains the
// session token.
const DefaultTokenHeaderName = "X-Ermes-Token"

// GetSessionTokenBytesFromHeader returns the session token bytes from the
// header of the request.
func GetSessionTokenBytesFromHeader(req *http.Request, headerName string) []byte {
	return []byte(req.Header.Get(headerName))
}

// SetSessionTokenBytesToHeader sets the session token bytes to the header of
// the response.
func SetSessionTokenBytesToHeader(w http.ResponseWriter, sessionTokenBytes []byte, headerName string) {
	w.Header().Set(headerName, string(sessionTokenBytes))
}
