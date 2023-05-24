package gen

import (
	"github.com/xlkness/lkit-go"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

type Method struct {
	Package     string
	ProtoMethod *descriptorpb.MethodDescriptorProto
}

func (m *Method) Name() string {
	return m.ProtoMethod.GetName()
}

func (m *Method) InputType() string {
	tokens := strings.Split(m.ProtoMethod.GetInputType(), ".")
	if len(tokens) < 3 {
		return m.ProtoMethod.GetInputType()
	}
	if tokens[1] == m.Package {
		return tokens[2]
	}
	return tokens[1] + "." + tokens[2]
}

func (m *Method) OutputType() string {
	tokens := strings.Split(m.ProtoMethod.GetOutputType(), ".")
	if len(tokens) < 3 {
		return m.ProtoMethod.GetOutputType()
	}
	if tokens[1] == m.Package {
		return tokens[2]
	}
	return tokens[1] + "." + tokens[2]
}

type Service struct {
	ProtoService *descriptorpb.ServiceDescriptorProto
	Methods      []*Method
}

func (s *Service) ServiceInterfaceName() string {
	return s.Name_FooBar() + "ServiceInterface"
}

func (s *Service) HandlerInterfaceName() string {
	return s.Name_FooBar() + "HandlerInterface"
}

func (s *Service) Name_FooBar() string {
	return lkit_go.StringCamelCase(s.ProtoService.GetName())
}

func (s *Service) Name_fooBar() string {
	return lkit_go.StringLowerCase(s.ProtoService.GetName())
}

type File struct {
	ProtoFile                    *descriptorpb.FileDescriptorProto
	Services                     []*Service
	IsEnableSpecInvokePeer       bool // 点对点调用
	IsEnableInvokeConsistentHash bool // 一致性hash调用
}

func (f *File) ServiceName_FooBar() string {
	return lkit_go.StringCamelCase(f.ServiceName())
}

func (f *File) ServiceName_fooBar() string {
	return lkit_go.StringLowerCase(f.ServiceName())
}

func (f *File) Package() string {
	return f.ProtoFile.GetPackage()
}

// Name 文件只定义了单个服务，就用这个服务名作名字，否则就用文件的基础名
func (f *File) ServiceName() string {
	if len(f.Services) <= 0 {
		return strings.Split(f.FileName(), ".")[0]
	}
	if len(f.Services) == 1 {
		return f.Services[0].ProtoService.GetName()
	}

	return strings.Split(f.FileName(), ".")[0]
}

func (f *File) FileName() string {
	return filepath.Base(f.ProtoFile.GetName())
}

func (f *File) BaseFileName() string {
	idx := strings.LastIndexByte(f.FileName(), byte('.'))
	if idx < 0 {
		return f.FileName()
	}
	return f.FileName()[:idx]
}

// func (f *File) FirstService() *Service {
// 	return f.Services[0]
// }

func (f *File) AllServices() []*Service {
	return f.Services
}

func (f *File) ServiceInterfaceName() string {
	if len(f.Services) > 1 {
		return f.ServiceName_FooBar() + "ServicesInterface"
	}
	return f.ServiceName_FooBar() + "ServiceInterface"
}

func (f *File) HandlerInterfaceName() string {
	if len(f.Services) > 1 {
		return f.ServiceName_FooBar() + "HandlersInterface"
	}
	return f.ServiceName_FooBar() + "HandlerInterface"
}

func (f *File) MultiServices() bool {
	return len(f.Services) > 1
}
