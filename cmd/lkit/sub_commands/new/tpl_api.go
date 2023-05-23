package new

var tplApiProto = `
syntax = "proto3";

package api_{{.AppName}};

// compile:protoc -I$PROJECT_PATH/src/proto -I. --gogofaster_out=. --joymicro_out=. *.proto

service {{.AppCamelName}} {
  rpc Say({{.AppCamelName}}Req) returns ({{.AppCamelName}}Res) {};
}

message {{.AppCamelName}}Req {
  string Name = 1;
  string Msg = 2;
}

message {{.AppCamelName}}Res {
  string Msg = 1;
}
`
