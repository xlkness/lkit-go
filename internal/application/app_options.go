package application

// WithAppBootFlag 设置app的起服参数，flags必须为结构体指针！
// 只支持string/int/int64/bool四种字段类型，例如：
//
//	type Flags struct {
//		F1 string `env:"id" desc:"boot id" default:"default value"`
//	  	F2 int `env:"num" desc:"number" default:"3"`
//	}
//	WithAppBootFlag(&Flags{})
func WithAppBootFlag(flag interface{}) AppOption {
	return appOptionFunction(func(app *Application) {
		app.bootFlag = flag
	})
}

type AppOption interface {
	Apply(scd *Application)
}

type appOptionFunction func(app *Application)

func (of appOptionFunction) Apply(app *Application) {
	of(app)
}
