package main

import (
	"fmt"
	"github.com/xlkness/lkit-go/cmd/protoc-gen-joymicro/gen"
	"github.com/xlkness/lkit-go/cmd/protoc-gen-joymicro/gen_all_in_one"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	mode := os.Getenv("joymicro_mode")
	str := ""
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		str += fmt.Sprint(err, "reading input")
	}

	request := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, request); err != nil {
		str += fmt.Sprint(err, "parsing input proto")
	}

	if len(request.FileToGenerate) == 0 {
		str += fmt.Sprint("no files to generate")
		os.Exit(1)
	}

	response := &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{},
	}

	if mode == "all_in_one" {
		for _, file := range request.ProtoFile {
			for _, f := range gen_all_in_one.GenerateFile(file) {
				if f != nil {
					response.File = append(response.File, f)
				}
			}
		}
	} else {
		for _, file := range request.ProtoFile {
			for _, f := range gen.GenerateFile(file) {
				if f != nil {
					response.File = append(response.File, f)
				}
			}
		}
	}

	data, err = proto.Marshal(response)
	if err != nil {
		str += fmt.Sprint(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		str += fmt.Sprint(err, "failed to write output proto")
	}
}
