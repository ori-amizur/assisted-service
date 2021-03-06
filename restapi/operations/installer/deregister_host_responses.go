// Code generated by go-swagger; DO NOT EDIT.

package installer

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openshift/assisted-service/models"
)

// DeregisterHostNoContentCode is the HTTP code returned for type DeregisterHostNoContent
const DeregisterHostNoContentCode int = 204

/*DeregisterHostNoContent Success.

swagger:response deregisterHostNoContent
*/
type DeregisterHostNoContent struct {
}

// NewDeregisterHostNoContent creates DeregisterHostNoContent with default headers values
func NewDeregisterHostNoContent() *DeregisterHostNoContent {

	return &DeregisterHostNoContent{}
}

// WriteResponse to the client
func (o *DeregisterHostNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// DeregisterHostBadRequestCode is the HTTP code returned for type DeregisterHostBadRequest
const DeregisterHostBadRequestCode int = 400

/*DeregisterHostBadRequest Error.

swagger:response deregisterHostBadRequest
*/
type DeregisterHostBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeregisterHostBadRequest creates DeregisterHostBadRequest with default headers values
func NewDeregisterHostBadRequest() *DeregisterHostBadRequest {

	return &DeregisterHostBadRequest{}
}

// WithPayload adds the payload to the deregister host bad request response
func (o *DeregisterHostBadRequest) WithPayload(payload *models.Error) *DeregisterHostBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the deregister host bad request response
func (o *DeregisterHostBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeregisterHostBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeregisterHostUnauthorizedCode is the HTTP code returned for type DeregisterHostUnauthorized
const DeregisterHostUnauthorizedCode int = 401

/*DeregisterHostUnauthorized Unauthorized.

swagger:response deregisterHostUnauthorized
*/
type DeregisterHostUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.InfraError `json:"body,omitempty"`
}

// NewDeregisterHostUnauthorized creates DeregisterHostUnauthorized with default headers values
func NewDeregisterHostUnauthorized() *DeregisterHostUnauthorized {

	return &DeregisterHostUnauthorized{}
}

// WithPayload adds the payload to the deregister host unauthorized response
func (o *DeregisterHostUnauthorized) WithPayload(payload *models.InfraError) *DeregisterHostUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the deregister host unauthorized response
func (o *DeregisterHostUnauthorized) SetPayload(payload *models.InfraError) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeregisterHostUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeregisterHostForbiddenCode is the HTTP code returned for type DeregisterHostForbidden
const DeregisterHostForbiddenCode int = 403

/*DeregisterHostForbidden Forbidden.

swagger:response deregisterHostForbidden
*/
type DeregisterHostForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.InfraError `json:"body,omitempty"`
}

// NewDeregisterHostForbidden creates DeregisterHostForbidden with default headers values
func NewDeregisterHostForbidden() *DeregisterHostForbidden {

	return &DeregisterHostForbidden{}
}

// WithPayload adds the payload to the deregister host forbidden response
func (o *DeregisterHostForbidden) WithPayload(payload *models.InfraError) *DeregisterHostForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the deregister host forbidden response
func (o *DeregisterHostForbidden) SetPayload(payload *models.InfraError) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeregisterHostForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeregisterHostNotFoundCode is the HTTP code returned for type DeregisterHostNotFound
const DeregisterHostNotFoundCode int = 404

/*DeregisterHostNotFound Error.

swagger:response deregisterHostNotFound
*/
type DeregisterHostNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeregisterHostNotFound creates DeregisterHostNotFound with default headers values
func NewDeregisterHostNotFound() *DeregisterHostNotFound {

	return &DeregisterHostNotFound{}
}

// WithPayload adds the payload to the deregister host not found response
func (o *DeregisterHostNotFound) WithPayload(payload *models.Error) *DeregisterHostNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the deregister host not found response
func (o *DeregisterHostNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeregisterHostNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeregisterHostMethodNotAllowedCode is the HTTP code returned for type DeregisterHostMethodNotAllowed
const DeregisterHostMethodNotAllowedCode int = 405

/*DeregisterHostMethodNotAllowed Method Not Allowed.

swagger:response deregisterHostMethodNotAllowed
*/
type DeregisterHostMethodNotAllowed struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeregisterHostMethodNotAllowed creates DeregisterHostMethodNotAllowed with default headers values
func NewDeregisterHostMethodNotAllowed() *DeregisterHostMethodNotAllowed {

	return &DeregisterHostMethodNotAllowed{}
}

// WithPayload adds the payload to the deregister host method not allowed response
func (o *DeregisterHostMethodNotAllowed) WithPayload(payload *models.Error) *DeregisterHostMethodNotAllowed {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the deregister host method not allowed response
func (o *DeregisterHostMethodNotAllowed) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeregisterHostMethodNotAllowed) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(405)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeregisterHostInternalServerErrorCode is the HTTP code returned for type DeregisterHostInternalServerError
const DeregisterHostInternalServerErrorCode int = 500

/*DeregisterHostInternalServerError Error.

swagger:response deregisterHostInternalServerError
*/
type DeregisterHostInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeregisterHostInternalServerError creates DeregisterHostInternalServerError with default headers values
func NewDeregisterHostInternalServerError() *DeregisterHostInternalServerError {

	return &DeregisterHostInternalServerError{}
}

// WithPayload adds the payload to the deregister host internal server error response
func (o *DeregisterHostInternalServerError) WithPayload(payload *models.Error) *DeregisterHostInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the deregister host internal server error response
func (o *DeregisterHostInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeregisterHostInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
