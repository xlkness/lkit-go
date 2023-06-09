package gen

var tempfileCommonText = `
package {{ .Package }}

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"context"
	"github.com/xlkness/lkit-go"
	"time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context

var _ = lkit_go.JoyService{}
`
