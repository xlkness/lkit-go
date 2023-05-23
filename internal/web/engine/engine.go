package engine

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type unmarshal func(data []byte, v any) error

// type HandlerFunc func(ctx Context)
type HandlerFunc interface{}

type RouterGroup struct {
	basePath         string
	group            *gin.RouterGroup
	GroupRoutes      map[string]*RouterGroup
	Routes           map[string]*RouteInfo
	newContextFun    func() Context
	UnmarshalHandler unmarshal
}

func newRouterGroup(group *gin.RouterGroup, newContextFun func() Context) *RouterGroup {
	return &RouterGroup{
		basePath:      group.BasePath(),
		group:         group,
		newContextFun: newContextFun,
		GroupRoutes:   make(map[string]*RouterGroup),
		Routes:        make(map[string]*RouteInfo),
	}
}

func (g *RouterGroup) SetUnmarshalHandler(handlerFunc unmarshal) *RouterGroup {
	g.UnmarshalHandler = handlerFunc
	return g
}

func (g *RouterGroup) Use(middleware ...HandlerFunc) {
	if len(middleware) == 0 {
		return
	}
	g.group.Use(getGinHandlerFun(g.newContextFun, nil, g.UnmarshalHandler, middleware...))
}

func (g *RouterGroup) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	grp := g.newGroup(path, handlers...)
	g.GroupRoutes[path] = grp
	return grp
}

func (g *RouterGroup) Get(path string, desc string, handlers ...HandlerFunc) gin.IRoutes {
	return g.GetWithStructParams(path, desc, nil, handlers...)
}

func (g *RouterGroup) GetWithStructParams(path string, desc string, structTemplate interface{}, handlers ...HandlerFunc) gin.IRoutes {
	g.Routes[path] = &RouteInfo{Desc: desc, Method: "GET", StructTemplate: structTemplate}
	if len(handlers) == 0 {
		return nil
	}
	return g.group.GET(path, getGinHandlerFun(g.newContextFun, structTemplate, g.UnmarshalHandler, handlers...))
}

func (g *RouterGroup) Post(path string, desc string, handlers ...HandlerFunc) gin.IRoutes {
	return g.PostWithStructParams(path, desc, nil, handlers)
}
func (g *RouterGroup) PostWithStructParams(path string, desc string, structTemplate interface{}, handlers ...HandlerFunc) gin.IRoutes {
	g.Routes[path] = &RouteInfo{Desc: desc, Method: "POST", StructTemplate: structTemplate}
	if len(handlers) == 0 {
		return nil
	}
	return g.group.POST(path, getGinHandlerFun(g.newContextFun, structTemplate, g.UnmarshalHandler, handlers...))
}

func (g *RouterGroup) TravelGroupTree() map[string]*RouteInfo {
	m := make(map[string]*RouteInfo)
	for k, route := range g.Routes {
		if k[0] != '/' {
			k = "/" + k
		}
		m[k] = route
	}
	for k, subG := range g.GroupRoutes {
		gm := subG.TravelGroupTree()
		for k1, v1 := range gm {
			if k1[0] != '/' {
				k1 = "/" + k1
			}
			m[k+k1] = v1
		}
	}
	return m
}

func (g *RouterGroup) newGroup(path string, handlers ...HandlerFunc) *RouterGroup {
	var grp *gin.RouterGroup
	if len(handlers) == 0 {
		grp = g.group.Group(path)
	} else {
		grp = g.group.Group(path, getGinHandlerFun(g.newContextFun, nil, g.UnmarshalHandler, handlers...))
	}

	grp1 := newRouterGroup(grp, g.newContextFun)
	return grp1
}

type Engine struct {
	Addr          string
	basePath      string
	ginEngine     *gin.Engine
	GroupRoutes   map[string]*RouterGroup // 组路由
	Routes        map[string]*RouteInfo   // 直接路由
	newContextFun func() Context
}

func NewEngine(addr string, newContextFun func() Context) *Engine {
	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())
	// ginEngine.Use(gin.Logger()) // 默认logger不可控，所有请求默认打印
	ginEngine.SetTrustedProxies([]string{addr})

	engine := &Engine{
		Addr:          addr,
		ginEngine:     ginEngine,
		newContextFun: newContextFun,
		GroupRoutes:   make(map[string]*RouterGroup),
		Routes:        make(map[string]*RouteInfo),
	}
	return engine
}

func (e *Engine) EnableDebugMode() {
	gin.SetMode(gin.DebugMode)
}

func (e *Engine) WithLogger(logger Logger) *Engine {
	e.Use(func(ctx Context) {
		params := parseContextLogParams(ctx)
		logger.HandleRequest(ctx, params)
	})
	return e
}

func (e *Engine) Use(middleware ...HandlerFunc) {
	e.ginEngine.Use(getGinHandlerFun(e.newContextFun, nil, nil, middleware...))
}

func (e *Engine) Group(path string, handlers ...HandlerFunc) *RouterGroup {
	grp := e.newGroup(path, handlers...)
	e.GroupRoutes[path] = grp
	return grp
}

func (e *Engine) Get(path string, desc string, handlers ...HandlerFunc) gin.IRoutes {
	return e.GetWithStructParams(path, desc, nil, handlers...)
}

func (e *Engine) GetWithStructParams(path string, desc string, structTemplate interface{}, handlers ...HandlerFunc) gin.IRoutes {
	e.Routes[path] = &RouteInfo{Desc: desc, Method: "GET", StructTemplate: structTemplate}
	return e.ginEngine.GET(path, getGinHandlerFun(e.newContextFun, structTemplate, nil, handlers...))
}

func (e *Engine) Post(path string, desc string, handlers ...HandlerFunc) gin.IRoutes {
	return e.PostWithStructParams(path, desc, nil, handlers...)
}

func (e *Engine) PostWithStructParams(path string, desc string, structTemplate interface{}, handlers ...HandlerFunc) gin.IRoutes {
	e.Routes[path] = &RouteInfo{Desc: desc, Method: "POST", StructTemplate: structTemplate}
	return e.ginEngine.POST(path, getGinHandlerFun(e.newContextFun, structTemplate, nil, handlers...))
}

func (e *Engine) newGroup(path string, handlers ...HandlerFunc) *RouterGroup {
	var grp *gin.RouterGroup
	if len(handlers) == 0 {
		grp = e.ginEngine.Group(path)
	} else {
		grp = e.ginEngine.Group(path, getGinHandlerFun(e.newContextFun, nil, nil, handlers...))
	}

	grp1 := newRouterGroup(grp, e.newContextFun)
	return grp1
}

func (e *Engine) TravelGroupTree() map[string]*RouteInfo {
	m := make(map[string]*RouteInfo)
	for k, route := range e.Routes {
		if k[0] != '/' {
			k = "/" + k
		}
		m[k] = route
	}
	for k, subG := range e.GroupRoutes {
		gm := subG.TravelGroupTree()
		for k1, v1 := range gm {
			if k1[0] != '/' {
				k1 = "/" + k1
			}
			m[k+k1] = v1
		}
	}
	return m
}

func (e *Engine) Run() error {
	return e.ginEngine.Run(e.Addr)
}

func (e *Engine) Stop() {

}

func (e *Engine) GetGinEngine() *gin.Engine {
	return e.ginEngine
}

func getGinHandlerFun(newContextFun func() Context, structTemplate interface{}, unmarshalHandler unmarshal, handlers ...HandlerFunc) gin.HandlerFunc {
	if len(handlers) == 0 {
		return nil
	}
	return func(c *gin.Context) {
		ctx := newContextFun()
		ctx.SetGinContext(c)
		for _, h := range handlers {
			if structTemplate != nil {
				uri, body, receiver, field, value, err := structuredUnmarshaler(c, unmarshalHandler, structTemplate)
				if err != nil {
					ctx.ResponseParseParamsFieldFail(c.FullPath(), uri, body, field, value, err)
					c.Abort()
					return
				} else {
					reflect.ValueOf(h).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(receiver)})
				}
			} else {
				reflect.ValueOf(h).Call([]reflect.Value{reflect.ValueOf(ctx)})
			}

			if c.IsAborted() {
				return
			}
		}
	}
}
