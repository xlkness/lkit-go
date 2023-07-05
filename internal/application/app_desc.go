package application

// ApplicationDescInfo 调度器创建app时注入的app描述信息
type ApplicationDescInfo struct {
	name     string
	initFunc func(globalBootFlag *CommBootFlag, globalBootFile interface{}, app *Application) error
	options  []AppOption
}

func NewApplicationDescInfo(name string, initFunc func(globalBootFlag *CommBootFlag, globalBootFile interface{}, app *Application) error) *ApplicationDescInfo {
	adi := new(ApplicationDescInfo)
	adi.name = name
	adi.initFunc = initFunc
	return adi
}

func (adi *ApplicationDescInfo) WithOptions(options ...AppOption) *ApplicationDescInfo {
	adi.options = append(adi.options, options...)
	return adi
}
