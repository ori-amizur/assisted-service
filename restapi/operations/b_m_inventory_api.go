// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/security"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/filanov/bm-inventory/restapi/operations/inventory"
)

// NewBMInventoryAPI creates a new BMInventory instance
func NewBMInventoryAPI(spec *loads.Document) *BMInventoryAPI {
	return &BMInventoryAPI{
		handlers:            make(map[string]map[string]http.Handler),
		formats:             strfmt.Default,
		defaultConsumes:     "application/json",
		defaultProduces:     "application/json",
		customConsumers:     make(map[string]runtime.Consumer),
		customProducers:     make(map[string]runtime.Producer),
		PreServerShutdown:   func() {},
		ServerShutdown:      func() {},
		spec:                spec,
		ServeError:          errors.ServeError,
		BasicAuthenticator:  security.BasicAuth,
		APIKeyAuthenticator: security.APIKeyAuth,
		BearerAuthenticator: security.BearerAuth,

		JSONConsumer: runtime.JSONConsumer(),

		BinProducer:  runtime.ByteStreamProducer(),
		JSONProducer: runtime.JSONProducer(),
		TextXYamlProducer: runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
			return errors.NotImplemented("textXYaml producer has not yet been implemented")
		}),

		InventoryDeregisterClusterHandler: inventory.DeregisterClusterHandlerFunc(func(params inventory.DeregisterClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.DeregisterCluster has not yet been implemented")
		}),
		InventoryDeregisterHostHandler: inventory.DeregisterHostHandlerFunc(func(params inventory.DeregisterHostParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.DeregisterHost has not yet been implemented")
		}),
		InventoryDisableHostHandler: inventory.DisableHostHandlerFunc(func(params inventory.DisableHostParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.DisableHost has not yet been implemented")
		}),
		InventoryDownloadClusterISOHandler: inventory.DownloadClusterISOHandlerFunc(func(params inventory.DownloadClusterISOParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.DownloadClusterISO has not yet been implemented")
		}),
		InventoryDownloadClusterKubeconfigHandler: inventory.DownloadClusterKubeconfigHandlerFunc(func(params inventory.DownloadClusterKubeconfigParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.DownloadClusterKubeconfig has not yet been implemented")
		}),
		InventoryEnableHostHandler: inventory.EnableHostHandlerFunc(func(params inventory.EnableHostParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.EnableHost has not yet been implemented")
		}),
		InventoryGetClusterHandler: inventory.GetClusterHandlerFunc(func(params inventory.GetClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.GetCluster has not yet been implemented")
		}),
		InventoryGetHostHandler: inventory.GetHostHandlerFunc(func(params inventory.GetHostParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.GetHost has not yet been implemented")
		}),
		InventoryGetNextStepsHandler: inventory.GetNextStepsHandlerFunc(func(params inventory.GetNextStepsParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.GetNextSteps has not yet been implemented")
		}),
		InventoryInstallClusterHandler: inventory.InstallClusterHandlerFunc(func(params inventory.InstallClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.InstallCluster has not yet been implemented")
		}),
		InventoryListClustersHandler: inventory.ListClustersHandlerFunc(func(params inventory.ListClustersParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.ListClusters has not yet been implemented")
		}),
		InventoryListHostsHandler: inventory.ListHostsHandlerFunc(func(params inventory.ListHostsParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.ListHosts has not yet been implemented")
		}),
		InventoryPostStepReplyHandler: inventory.PostStepReplyHandlerFunc(func(params inventory.PostStepReplyParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.PostStepReply has not yet been implemented")
		}),
		InventoryRegisterClusterHandler: inventory.RegisterClusterHandlerFunc(func(params inventory.RegisterClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.RegisterCluster has not yet been implemented")
		}),
		InventoryRegisterHostHandler: inventory.RegisterHostHandlerFunc(func(params inventory.RegisterHostParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.RegisterHost has not yet been implemented")
		}),
		InventorySetDebugStepHandler: inventory.SetDebugStepHandlerFunc(func(params inventory.SetDebugStepParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.SetDebugStep has not yet been implemented")
		}),
		InventoryUpdateClusterHandler: inventory.UpdateClusterHandlerFunc(func(params inventory.UpdateClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation inventory.UpdateCluster has not yet been implemented")
		}),
	}
}

/*BMInventoryAPI Bare metal inventory */
type BMInventoryAPI struct {
	spec            *loads.Document
	context         *middleware.Context
	handlers        map[string]map[string]http.Handler
	formats         strfmt.Registry
	customConsumers map[string]runtime.Consumer
	customProducers map[string]runtime.Producer
	defaultConsumes string
	defaultProduces string
	Middleware      func(middleware.Builder) http.Handler

	// BasicAuthenticator generates a runtime.Authenticator from the supplied basic auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	BasicAuthenticator func(security.UserPassAuthentication) runtime.Authenticator
	// APIKeyAuthenticator generates a runtime.Authenticator from the supplied token auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	APIKeyAuthenticator func(string, string, security.TokenAuthentication) runtime.Authenticator
	// BearerAuthenticator generates a runtime.Authenticator from the supplied bearer token auth function.
	// It has a default implementation in the security package, however you can replace it for your particular usage.
	BearerAuthenticator func(string, security.ScopedTokenAuthentication) runtime.Authenticator

	// JSONConsumer registers a consumer for the following mime types:
	//   - application/json
	JSONConsumer runtime.Consumer

	// BinProducer registers a producer for the following mime types:
	//   - application/octet-stream
	BinProducer runtime.Producer
	// JSONProducer registers a producer for the following mime types:
	//   - application/json
	JSONProducer runtime.Producer
	// TextXYamlProducer registers a producer for the following mime types:
	//   - text/x-yaml
	TextXYamlProducer runtime.Producer

	// InventoryDeregisterClusterHandler sets the operation handler for the deregister cluster operation
	InventoryDeregisterClusterHandler inventory.DeregisterClusterHandler
	// InventoryDeregisterHostHandler sets the operation handler for the deregister host operation
	InventoryDeregisterHostHandler inventory.DeregisterHostHandler
	// InventoryDisableHostHandler sets the operation handler for the disable host operation
	InventoryDisableHostHandler inventory.DisableHostHandler
	// InventoryDownloadClusterISOHandler sets the operation handler for the download cluster i s o operation
	InventoryDownloadClusterISOHandler inventory.DownloadClusterISOHandler
	// InventoryDownloadClusterKubeconfigHandler sets the operation handler for the download cluster kubeconfig operation
	InventoryDownloadClusterKubeconfigHandler inventory.DownloadClusterKubeconfigHandler
	// InventoryEnableHostHandler sets the operation handler for the enable host operation
	InventoryEnableHostHandler inventory.EnableHostHandler
	// InventoryGetClusterHandler sets the operation handler for the get cluster operation
	InventoryGetClusterHandler inventory.GetClusterHandler
	// InventoryGetHostHandler sets the operation handler for the get host operation
	InventoryGetHostHandler inventory.GetHostHandler
	// InventoryGetNextStepsHandler sets the operation handler for the get next steps operation
	InventoryGetNextStepsHandler inventory.GetNextStepsHandler
	// InventoryInstallClusterHandler sets the operation handler for the install cluster operation
	InventoryInstallClusterHandler inventory.InstallClusterHandler
	// InventoryListClustersHandler sets the operation handler for the list clusters operation
	InventoryListClustersHandler inventory.ListClustersHandler
	// InventoryListHostsHandler sets the operation handler for the list hosts operation
	InventoryListHostsHandler inventory.ListHostsHandler
	// InventoryPostStepReplyHandler sets the operation handler for the post step reply operation
	InventoryPostStepReplyHandler inventory.PostStepReplyHandler
	// InventoryRegisterClusterHandler sets the operation handler for the register cluster operation
	InventoryRegisterClusterHandler inventory.RegisterClusterHandler
	// InventoryRegisterHostHandler sets the operation handler for the register host operation
	InventoryRegisterHostHandler inventory.RegisterHostHandler
	// InventorySetDebugStepHandler sets the operation handler for the set debug step operation
	InventorySetDebugStepHandler inventory.SetDebugStepHandler
	// InventoryUpdateClusterHandler sets the operation handler for the update cluster operation
	InventoryUpdateClusterHandler inventory.UpdateClusterHandler
	// ServeError is called when an error is received, there is a default handler
	// but you can set your own with this
	ServeError func(http.ResponseWriter, *http.Request, error)

	// PreServerShutdown is called before the HTTP(S) server is shutdown
	// This allows for custom functions to get executed before the HTTP(S) server stops accepting traffic
	PreServerShutdown func()

	// ServerShutdown is called when the HTTP(S) server is shut down and done
	// handling all active connections and does not accept connections any more
	ServerShutdown func()

	// Custom command line argument groups with their descriptions
	CommandLineOptionsGroups []swag.CommandLineOptionsGroup

	// User defined logger function.
	Logger func(string, ...interface{})
}

// SetDefaultProduces sets the default produces media type
func (o *BMInventoryAPI) SetDefaultProduces(mediaType string) {
	o.defaultProduces = mediaType
}

// SetDefaultConsumes returns the default consumes media type
func (o *BMInventoryAPI) SetDefaultConsumes(mediaType string) {
	o.defaultConsumes = mediaType
}

// SetSpec sets a spec that will be served for the clients.
func (o *BMInventoryAPI) SetSpec(spec *loads.Document) {
	o.spec = spec
}

// DefaultProduces returns the default produces media type
func (o *BMInventoryAPI) DefaultProduces() string {
	return o.defaultProduces
}

// DefaultConsumes returns the default consumes media type
func (o *BMInventoryAPI) DefaultConsumes() string {
	return o.defaultConsumes
}

// Formats returns the registered string formats
func (o *BMInventoryAPI) Formats() strfmt.Registry {
	return o.formats
}

// RegisterFormat registers a custom format validator
func (o *BMInventoryAPI) RegisterFormat(name string, format strfmt.Format, validator strfmt.Validator) {
	o.formats.Add(name, format, validator)
}

// Validate validates the registrations in the BMInventoryAPI
func (o *BMInventoryAPI) Validate() error {
	var unregistered []string

	if o.JSONConsumer == nil {
		unregistered = append(unregistered, "JSONConsumer")
	}

	if o.BinProducer == nil {
		unregistered = append(unregistered, "BinProducer")
	}
	if o.JSONProducer == nil {
		unregistered = append(unregistered, "JSONProducer")
	}
	if o.TextXYamlProducer == nil {
		unregistered = append(unregistered, "TextXYamlProducer")
	}

	if o.InventoryDeregisterClusterHandler == nil {
		unregistered = append(unregistered, "inventory.DeregisterClusterHandler")
	}
	if o.InventoryDeregisterHostHandler == nil {
		unregistered = append(unregistered, "inventory.DeregisterHostHandler")
	}
	if o.InventoryDisableHostHandler == nil {
		unregistered = append(unregistered, "inventory.DisableHostHandler")
	}
	if o.InventoryDownloadClusterISOHandler == nil {
		unregistered = append(unregistered, "inventory.DownloadClusterISOHandler")
	}
	if o.InventoryDownloadClusterKubeconfigHandler == nil {
		unregistered = append(unregistered, "inventory.DownloadClusterKubeconfigHandler")
	}
	if o.InventoryEnableHostHandler == nil {
		unregistered = append(unregistered, "inventory.EnableHostHandler")
	}
	if o.InventoryGetClusterHandler == nil {
		unregistered = append(unregistered, "inventory.GetClusterHandler")
	}
	if o.InventoryGetHostHandler == nil {
		unregistered = append(unregistered, "inventory.GetHostHandler")
	}
	if o.InventoryGetNextStepsHandler == nil {
		unregistered = append(unregistered, "inventory.GetNextStepsHandler")
	}
	if o.InventoryInstallClusterHandler == nil {
		unregistered = append(unregistered, "inventory.InstallClusterHandler")
	}
	if o.InventoryListClustersHandler == nil {
		unregistered = append(unregistered, "inventory.ListClustersHandler")
	}
	if o.InventoryListHostsHandler == nil {
		unregistered = append(unregistered, "inventory.ListHostsHandler")
	}
	if o.InventoryPostStepReplyHandler == nil {
		unregistered = append(unregistered, "inventory.PostStepReplyHandler")
	}
	if o.InventoryRegisterClusterHandler == nil {
		unregistered = append(unregistered, "inventory.RegisterClusterHandler")
	}
	if o.InventoryRegisterHostHandler == nil {
		unregistered = append(unregistered, "inventory.RegisterHostHandler")
	}
	if o.InventorySetDebugStepHandler == nil {
		unregistered = append(unregistered, "inventory.SetDebugStepHandler")
	}
	if o.InventoryUpdateClusterHandler == nil {
		unregistered = append(unregistered, "inventory.UpdateClusterHandler")
	}

	if len(unregistered) > 0 {
		return fmt.Errorf("missing registration: %s", strings.Join(unregistered, ", "))
	}

	return nil
}

// ServeErrorFor gets a error handler for a given operation id
func (o *BMInventoryAPI) ServeErrorFor(operationID string) func(http.ResponseWriter, *http.Request, error) {
	return o.ServeError
}

// AuthenticatorsFor gets the authenticators for the specified security schemes
func (o *BMInventoryAPI) AuthenticatorsFor(schemes map[string]spec.SecurityScheme) map[string]runtime.Authenticator {
	return nil
}

// Authorizer returns the registered authorizer
func (o *BMInventoryAPI) Authorizer() runtime.Authorizer {
	return nil
}

// ConsumersFor gets the consumers for the specified media types.
// MIME type parameters are ignored here.
func (o *BMInventoryAPI) ConsumersFor(mediaTypes []string) map[string]runtime.Consumer {
	result := make(map[string]runtime.Consumer, len(mediaTypes))
	for _, mt := range mediaTypes {
		switch mt {
		case "application/json":
			result["application/json"] = o.JSONConsumer
		}

		if c, ok := o.customConsumers[mt]; ok {
			result[mt] = c
		}
	}
	return result
}

// ProducersFor gets the producers for the specified media types.
// MIME type parameters are ignored here.
func (o *BMInventoryAPI) ProducersFor(mediaTypes []string) map[string]runtime.Producer {
	result := make(map[string]runtime.Producer, len(mediaTypes))
	for _, mt := range mediaTypes {
		switch mt {
		case "application/octet-stream":
			result["application/octet-stream"] = o.BinProducer
		case "application/json":
			result["application/json"] = o.JSONProducer
		case "text/x-yaml":
			result["text/x-yaml"] = o.TextXYamlProducer
		}

		if p, ok := o.customProducers[mt]; ok {
			result[mt] = p
		}
	}
	return result
}

// HandlerFor gets a http.Handler for the provided operation method and path
func (o *BMInventoryAPI) HandlerFor(method, path string) (http.Handler, bool) {
	if o.handlers == nil {
		return nil, false
	}
	um := strings.ToUpper(method)
	if _, ok := o.handlers[um]; !ok {
		return nil, false
	}
	if path == "/" {
		path = ""
	}
	h, ok := o.handlers[um][path]
	return h, ok
}

// Context returns the middleware context for the b m inventory API
func (o *BMInventoryAPI) Context() *middleware.Context {
	if o.context == nil {
		o.context = middleware.NewRoutableContext(o.spec, o, nil)
	}

	return o.context
}

func (o *BMInventoryAPI) initHandlerCache() {
	o.Context() // don't care about the result, just that the initialization happened
	if o.handlers == nil {
		o.handlers = make(map[string]map[string]http.Handler)
	}

	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/clusters/{clusterId}"] = inventory.NewDeregisterCluster(o.context, o.InventoryDeregisterClusterHandler)
	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/clusters/{clusterId}/hosts/{hostId}"] = inventory.NewDeregisterHost(o.context, o.InventoryDeregisterHostHandler)
	if o.handlers["DELETE"] == nil {
		o.handlers["DELETE"] = make(map[string]http.Handler)
	}
	o.handlers["DELETE"]["/clusters/{clusterId}/hosts/{hostId}/actions/enable"] = inventory.NewDisableHost(o.context, o.InventoryDisableHostHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters/{clusterId}/downloads/image"] = inventory.NewDownloadClusterISO(o.context, o.InventoryDownloadClusterISOHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters/{clusterId}/downloads/kubeconfig"] = inventory.NewDownloadClusterKubeconfig(o.context, o.InventoryDownloadClusterKubeconfigHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/clusters/{clusterId}/hosts/{hostId}/actions/enable"] = inventory.NewEnableHost(o.context, o.InventoryEnableHostHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters/{clusterId}"] = inventory.NewGetCluster(o.context, o.InventoryGetClusterHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters/{clusterId}/hosts/{hostId}"] = inventory.NewGetHost(o.context, o.InventoryGetHostHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters/{clusterId}/hosts/{hostId}/instructions"] = inventory.NewGetNextSteps(o.context, o.InventoryGetNextStepsHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/clusters/{clusterId}/actions/install"] = inventory.NewInstallCluster(o.context, o.InventoryInstallClusterHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters"] = inventory.NewListClusters(o.context, o.InventoryListClustersHandler)
	if o.handlers["GET"] == nil {
		o.handlers["GET"] = make(map[string]http.Handler)
	}
	o.handlers["GET"]["/clusters/{clusterId}/hosts"] = inventory.NewListHosts(o.context, o.InventoryListHostsHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/clusters/{clusterId}/hosts/{hostId}/instructions"] = inventory.NewPostStepReply(o.context, o.InventoryPostStepReplyHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/clusters"] = inventory.NewRegisterCluster(o.context, o.InventoryRegisterClusterHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/clusters/{clusterId}/hosts"] = inventory.NewRegisterHost(o.context, o.InventoryRegisterHostHandler)
	if o.handlers["POST"] == nil {
		o.handlers["POST"] = make(map[string]http.Handler)
	}
	o.handlers["POST"]["/clusters/{clusterId}/hosts/{hostId}/actions/debug"] = inventory.NewSetDebugStep(o.context, o.InventorySetDebugStepHandler)
	if o.handlers["PATCH"] == nil {
		o.handlers["PATCH"] = make(map[string]http.Handler)
	}
	o.handlers["PATCH"]["/clusters/{clusterId}"] = inventory.NewUpdateCluster(o.context, o.InventoryUpdateClusterHandler)
}

// Serve creates a http handler to serve the API over HTTP
// can be used directly in http.ListenAndServe(":8000", api.Serve(nil))
func (o *BMInventoryAPI) Serve(builder middleware.Builder) http.Handler {
	o.Init()

	if o.Middleware != nil {
		return o.Middleware(builder)
	}
	return o.context.APIHandler(builder)
}

// Init allows you to just initialize the handler cache, you can then recompose the middleware as you see fit
func (o *BMInventoryAPI) Init() {
	if len(o.handlers) == 0 {
		o.initHandlerCache()
	}
}

// RegisterConsumer allows you to add (or override) a consumer for a media type.
func (o *BMInventoryAPI) RegisterConsumer(mediaType string, consumer runtime.Consumer) {
	o.customConsumers[mediaType] = consumer
}

// RegisterProducer allows you to add (or override) a producer for a media type.
func (o *BMInventoryAPI) RegisterProducer(mediaType string, producer runtime.Producer) {
	o.customProducers[mediaType] = producer
}

// AddMiddlewareFor adds a http middleware to existing handler
func (o *BMInventoryAPI) AddMiddlewareFor(method, path string, builder middleware.Builder) {
	um := strings.ToUpper(method)
	if path == "/" {
		path = ""
	}
	o.Init()
	if h, ok := o.handlers[um][path]; ok {
		o.handlers[method][path] = builder(h)
	}
}
