package gen_all_in_one

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var PluginName = "protoc-gen-joymicro"

func GenerateFile(protoFile *descriptorpb.FileDescriptorProto) []*pluginpb.CodeGeneratorResponse_File {
	file := &File{
		ProtoFile: protoFile,
	}
	for _, service := range protoFile.Service {
		s := &Service{
			ProtoService: service,
		}
		for _, m := range s.ProtoService.GetMethod() {
			s.Methods = append(s.Methods, &Method{ProtoMethod: m, Package: file.Package()})
		}
		file.Services = append(file.Services, s)
	}

	for _, loc := range file.ProtoFile.SourceCodeInfo.GetLocation() {
		c1 := loc.GetTrailingComments()
		c2 := loc.GetLeadingDetachedComments()
		c3 := loc.GetLeadingComments()
		checkF := func(c string) {
			if strings.Contains(c, "EnablePeer2Peer") {
				file.IsEnableSpecInvokePeer = true
			}
			if strings.Contains(c, "EnableConsistentHash") {
				file.IsEnableInvokeConsistentHash = true
			}
		}
		checkF(c1)
		for _, v := range c2 {
			checkF(v)
		}
		checkF(c3)
	}

	return append([]*pluginpb.CodeGeneratorResponse_File{generateFileServices(file)})
}

func generateFileServices(file *File) *pluginpb.CodeGeneratorResponse_File {
	if len(file.Services) <= 0 {
		return nil
	}

	outputFileName := file.BaseFileName() + ".joymicro.pb.go"

	var text = tempServicesText

	tmpl, err := template.New("joymicro.services").Parse(text)
	if err != nil {
		panic(err)
	}

	header := fmt.Sprintf("// Code generated by %s. DO NOT EDIT.\n", PluginName)
	header += fmt.Sprintf("// source: %s\n\n", file.FileName())

	buf := bytes.NewBuffer([]byte(header))
	err = tmpl.Execute(buf, file)
	if err != nil {
		panic(err)
	}

	newBuf, err := sortImports(buf.Bytes())
	if err != nil {
		panic(err)
	}

	fileOutPut := &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(outputFileName),
		Content: proto.String(string(newBuf)),
	}

	return fileOutPut
}

func generateFileApis(file *File) *pluginpb.CodeGeneratorResponse_File {
	if len(file.Services) <= 0 {
		return nil
	}

	outputFileName := "api_" + file.BaseFileName() + ".joymicro.pb.go"

	var text = `tempApiText`

	tmpl, err := template.New("joymicro.api").Parse(text)
	if err != nil {
		panic(err)
	}

	header := fmt.Sprintf("// Code generated by %s. DO NOT EDIT.\n", PluginName)
	header += fmt.Sprintf("// source: %s\n\n", file.FileName())

	buf := bytes.NewBuffer([]byte(header))
	err = tmpl.Execute(buf, file)
	if err != nil {
		panic(err)
	}

	newBuf, err := sortImports(buf.Bytes())
	if err != nil {
		panic(err)
	}

	fileOutPut := &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(outputFileName),
		Content: proto.String(string(newBuf)),
	}

	return fileOutPut
}

func sortImports(content []byte) ([]byte, error) {
	fset := token.NewFileSet()
	original := content
	fileAST, err := parser.ParseFile(fset, "", original, parser.ParseComments)
	if err != nil {
		// Print out the bad code with line numbers.
		// This should never happen in practice, but it can while changing generated code,
		// so consider this a debugging aid.
		var src bytes.Buffer
		s := bufio.NewScanner(bytes.NewReader(original))
		for line := 1; s.Scan(); line++ {
			fmt.Fprintf(&src, "%5d\t%s\n", line, s.Bytes())
		}
		return content, fmt.Errorf("bad Go source code was generated:%v, %v", err.Error(), "\n"+src.String())
	}
	ast.SortImports(fset, fileAST)

	buf := bytes.NewBuffer(nil)
	(&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(buf, fset, fileAST)
	return buf.Bytes(), nil
}