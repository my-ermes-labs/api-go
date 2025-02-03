package http

import (
	"context"
	"net/http"

	"github.com/my-ermes-labs/api-go/api"
	// "github.com/my-ermes-labs/log"
)

func CreateHandler(
	n *api.Node,
	opt HandlerOptions,
	handler func(w http.ResponseWriter, req *http.Request, sessionToken api.SessionToken) error,
) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		Handle(n, w, req, opt, handler)
	}
}

// This function handle the full lifecycle of a request, and allow to provide a
// callback that will be run if the session is acquired. The options allow to
// customize the behavior of the function.
//
// There are 3 possible outcomes:
//  1. The session is acquired and the callback is run.
//     1.1. The callback returns an error and the error response is returned.
//     (Note that the error should be returned before writing anything to
//     the responseWriter).
//     1.2. The callback returns nil and the response is returned.
//  2. The session has been offloaded and the callback is not run.
//  3. There is an error and the callback is not run.
func Handle(
	n *api.Node,
	w http.ResponseWriter,
	req *http.Request,
	opt HandlerOptions,
	handler func(w http.ResponseWriter, req *http.Request, sessionToken api.SessionToken) error) {

	// log.MyLog("START HANDLEEEEEEE")

	// Try to get the session token from the request.
	sessionTokenBytes := opt.getSessionTokenBytes(req)
	// log.MyLog("sessionTokenBytes = " + string(sessionTokenBytes))

	// If there is a session token and it belongs to a dummy client that ws not
	sessionToken, err := api.UnmarshallSessionToken(sessionTokenBytes)
	// log("TOKEN: \nHost = " + sessionToken.Host + "\nId = " + sessionToken.SessionId + "\nLocation Host = " + sessionToken.SessionLocation.Host + "\nLocation Id = " + sessionToken.SessionLocation.SessionId)

	// If there is an error, return an error response.
	if err != nil {
		opt.malformedSessionTokenErrorResponse(w, err)
		return
	}

	// If there is a session token and it belongs to a dummy client that was not
	// able to make the request to the correct node, redirect the request to the
	// correct node.
	if sessionToken != nil {
		// log.MyLog("Session Token != NULL")
		if redirect, destination := dummyClientNeedsRedirect(n, req.Context(), sessionToken); redirect {
			// Set the session sessionToken in the response.
			opt.setSessionTokenBytes(w, sessionTokenBytes)
			// Create the redirect response.
			opt.redirectResponse(w, req, destination.Host)
			// Return.
			return
		}
	}

	// If the client does not already have a session.
	if sessionToken == nil {
		// o := opt.redirectNewRequest(req, n)
		// log.MyLog("Session Token == NULL; redirect = " + strconv.FormatBool(o))
		// If the node must redirect new requests, redirect the request.
		if opt.redirectNewRequest(req, n) {
			// log.MyLog("Redirect")
			// Get the host to redirect the request to.
			host := opt.redirectTarget(req, n)
			// log.MyLog("host = " + host)
			// Create the redirect response.
			opt.redirectResponse(w, req, host)
			// Return.
			return
		}
	}

	if sessionToken == nil {
		// log.MyLog("Session Token == NULL --> create and acquire session")
		// Create a new session and acquire it to run the handler callback,
		// then update the session token.
		_, err = n.CreateAndAcquireSession(
			// Use the request context.
			req.Context(),
			// Create the options.
			api.CreateAndAcquireSessionOptions{
				CreateSessionOptions:  opt.getCreateSessionOptions(req),
				AcquireSessionOptions: opt.getAcquireSessionOptions(req),
			},
			// Wrap the handler callback.
			func(sessionToken api.SessionToken) error {
				sessionTokenBytes, err = api.MarshallSessionToken(sessionToken)
				// log.MyLog("If created and acquired \nSession Token -->Id = " + sessionToken.SessionId + "\nHost= " + sessionToken.Host)
				// It should not happen, but if there is an error, panic.
				if err != nil {
					panic(err)
				}
				// Set the session sessionToken in the response.
				opt.setSessionTokenBytes(w, sessionTokenBytes)
				// Run the handler callback.
				return handler(w, req, sessionToken)
			})

	} else {
		var newToken *api.SessionToken = nil
		// Acquire the session.
		newToken, err = n.AcquireSession(
			// Use the request context.
			req.Context(),
			// Pass the session token.
			*sessionToken,
			// Create the options.
			opt.getAcquireSessionOptions(req),
			// Wrap the handler callback.
			func() error {
				return handler(w, req, *sessionToken)
			})

		// If the session has been offloaded, redirect the request.
		if err == nil && newToken != nil {
			// Set the session token in the response.
			opt.setSessionTokenBytes(w, sessionTokenBytes)
			// Create the redirect response.
			opt.redirectResponse(w, req, sessionToken.Host)
		}
	}

	// If there is an error, return an error response.
	if err != nil {
		// Create the internal server error response.
		opt.internalServerErrorResponse(w, err)
		// Return.
		return
	}
}

// Return if the session token belongs to a dummy client that was not able to
// make the request to the correct node, and the sessionLocation of the correct node.
func dummyClientNeedsRedirect(n *api.Node, ctx context.Context, sessionToken *api.SessionToken) (bool, api.SessionLocation) {
	return sessionToken.Host != n.Host, sessionToken.SessionLocation
}
