// Code generated by go-swagger; DO NOT EDIT.

package inventory

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetNextStepsHandlerFunc turns a function with the right signature into a get next steps handler
type GetNextStepsHandlerFunc func(GetNextStepsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetNextStepsHandlerFunc) Handle(params GetNextStepsParams) middleware.Responder {
	return fn(params)
}

// GetNextStepsHandler interface for that can handle valid get next steps params
type GetNextStepsHandler interface {
	Handle(GetNextStepsParams) middleware.Responder
}

// NewGetNextSteps creates a new http.Handler for the get next steps operation
func NewGetNextSteps(ctx *middleware.Context, handler GetNextStepsHandler) *GetNextSteps {
	return &GetNextSteps{Context: ctx, Handler: handler}
}

/*GetNextSteps swagger:route GET /clusters/{clusterId}/hosts/{hostId}/instructions inventory getNextSteps

Retrieve the next operations that the agent need to perform

*/
type GetNextSteps struct {
	Context *middleware.Context
	Handler GetNextStepsHandler
}

func (o *GetNextSteps) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetNextStepsParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
