package lkit_go

import "github.com/xlkness/lkit-go/internal/web/engine"

type WebEngineContext = engine.Context
type WebEngine = engine.Engine

func NewEngine(addr string, newContextFun func() WebEngineContext) *WebEngine {
	return engine.NewEngine(addr, newContextFun)
}
