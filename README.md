# Service locator pattern

Simple (100 LoC) implementation of the [service locator pattern](https://en.wikipedia.org/wiki/Service_locator_pattern)
in Go.

## Usage

```go
package main

import (
    "net/http"
    "github.com/gabrielhora/locator"
)

func main() {
    svc := locator.New()

    /* Register services */

    // Singleton will always return the same instance when requested
    svc.Singleton("MySingletonService", func(loc locator.Locator) interface{} {
        // todo: return instance of service
        return &SingletonService{}
    })

    // Transient will create a new service everytime you request an instance
    svc.Transient("MyTransientService", func(loc locator.Locator) interface{} {
        // todo: return instance of service
        return &TransientService{}
    })

    // Scoped will return the same instance during the same request
    // For this to work you need to use `svc.Middleware` in your http router
    svc.Scoped("MyScopedService", func(req *http.Request, loc locator.Locator) interface{} {
        // todo: return instance of service

        // you can use `loc` to request other required services
        return &ScopedService{
            mySingleton: loc.Get("MySingletonService").(*SingletonService),
        }
    })

    /* Request instances */

    svc.Get("MySingletonService").(*SingletonService)
    
    svc.Get("MyTransientService").(*TransientService)

    // to get request scoped services you need to provide the request Context
    svc.GetScoped(req.Context(), "MyScopedService").(*ScopedService)
}
```
