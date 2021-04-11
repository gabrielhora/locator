package locator

import (
	"context"
	"log"
	"net/http"
)

type Locator interface {
	Get(name string) interface{}
	GetScoped(ctx context.Context, name string) interface{}
}

type ServiceBuilder func(loc Locator) interface{}
type ScopedServiceBuilder func(req *http.Request, loc Locator) interface{}

type DefaultLocator struct {
	services map[string]*service
}

type service struct {
	// either a ServiceBuilder or ScopedServiceBuilder
	builder   interface{}
	singleton bool
	scoped    bool
	instance  interface{}
}

var _ Locator = (*DefaultLocator)(nil)

func (loc *DefaultLocator) Get(name string) interface{} {
	srv, ok := loc.services[name]
	if !ok {
		log.Fatalf("could not find service %q", name)
	}

	if srv.singleton {
		if srv.instance == nil {
			srv.instance = srv.builder.(ServiceBuilder)(loc)
		}
		return srv.instance
	}

	if srv.scoped {
		log.Fatalf("use GetScoped to get scoped services")
	}

	return srv.builder.(ServiceBuilder)(loc)
}

func (loc *DefaultLocator) GetScoped(ctx context.Context, name string) interface{} {
	srv, ok := loc.services[name]
	if !ok {
		log.Fatalf("could not find service %q", name)
	}
	if !srv.scoped {
		log.Fatalf("GetScoped should be used only for Scoped services")
	}
	return ctx.Value(name)
}

func New() *DefaultLocator {
	return &DefaultLocator{services: map[string]*service{}}
}

// Middleware builds all request scoped services and add them to the request Context
func (loc *DefaultLocator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// build all request scoped services and add them to the request context
		for name, srv := range loc.services {
			if srv.scoped {
				ctx = context.WithValue(ctx, name, srv.builder.(ScopedServiceBuilder)(r, loc))
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Transient registers a transient dependency, which builds new instances at every request
func (loc *DefaultLocator) Transient(name string, builder ServiceBuilder) {
	loc.services[name] = &service{
		builder:   builder,
		singleton: false,
	}
}

// Scoped registers a request scoped service. For this to work register the Middleware in your router
func (loc *DefaultLocator) Scoped(name string, builder ScopedServiceBuilder) {
	loc.services[name] = &service{
		builder: builder,
		scoped:  true,
	}
}

// Singleton registers a singleton service. Do not resolve scoped services from a singleton
// and be careful not to do so indirectly, for example, through a transient service
func (loc *DefaultLocator) Singleton(name string, builder ServiceBuilder) {
	loc.services[name] = &service{
		builder:   builder,
		singleton: true,
	}
}
