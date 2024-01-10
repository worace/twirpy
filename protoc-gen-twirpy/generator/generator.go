package generator

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/proto"
)

func Generate(r *plugin.CodeGeneratorRequest) *plugin.CodeGeneratorResponse {
	resp := &plugin.CodeGeneratorResponse{}
	resp.SupportedFeatures = proto.Uint64(uint64(plugin.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))

	files := r.GetFileToGenerate()
	for _, fileName := range files {
		fd, err := getFileDescriptor(r.GetProtoFile(), fileName)
		if err != nil {
			resp.Error = proto.String("File[" + fileName + "][descriptor]: " + err.Error())
			return resp
		}

		twirpFile, err := GenerateTwirpFile(fd, fileName)
		if err != nil {
			resp.Error = proto.String("File[" + fileName + "][generate]: " + err.Error())
			return resp
		}
		resp.File = append(resp.File, twirpFile)
	}
	return resp
}

func GenerateTwirpFile(fd *descriptor.FileDescriptorProto, sourceFileName string) (*plugin.CodeGeneratorResponse_File, error) {

	name := fd.GetName()
	l := log.New(os.Stderr, "", 0)
	l.Println("Generating twirp file for", name)
	// package: haberdasher
	// fd.name: services/haberdasher.proto
	// sourceFileName: services/haberdasher.proto
	// destination file name: haberdasher_twirp.py

	l.Println("fd package: ", fd.GetPackage())
	l.Println("src filename: ", sourceFileName)
	// Content for schemas are put in a separate file with the same name as the proto file
	// but with _pb2.py suffix, e.g. haberdasher.proto -> haberdasher_pb2.py
	schema_src_name := strings.TrimSuffix(name, path.Ext(name)) + "_pb2.py"

	l.Println("schema filename: ", schema_src_name)
	vars := TwirpTemplateVariables{
		FileName:       name,
		SchemaFileName: schema_src_name,
	}

	svcs := fd.GetService()
	for _, svc := range svcs {
		// package: haberdasher
		// service name: Haberdasher
		l.Println("adding service: ", svc.GetName())
		svcURL := fmt.Sprintf("%s.%s", fd.GetPackage(), svc.GetName())
		twirpSvc := &TwirpService{
			Name:       svc.GetName(),
			ServiceURL: svcURL,
		}

		for _, method := range svc.GetMethod() {
			method.GetInputType()

			l.Println("adding method input: ", method.GetInputType())
			inputMessageNameComponents := strings.Split(method.GetInputType(), ".")
			lastInputMessageNameElement := inputMessageNameComponents[len(inputMessageNameComponents)-1]
			nextToLastInputMessageNameElement := inputMessageNameComponents[len(inputMessageNameComponents)-2]
			outputMessageNameComponents := strings.Split(method.GetOutputType(), ".")
			outputMessageName := outputMessageNameComponents[len(outputMessageNameComponents)-1]
			outputModuleName := outputMessageNameComponents[len(outputMessageNameComponents)-2]

			twirpMethod := &TwirpMethod{
				ServiceURL:              svcURL,
				ServiceName:             twirpSvc.Name,
				Name:                    method.GetName(),
				Input:                   getSymbol(method.GetInputType()),
				InputMessageName:        lastInputMessageNameElement,
				InputMessageModuleName:  nextToLastInputMessageNameElement,
				OutputMessageName:       outputMessageName,
				OutputMessageModuleName: outputModuleName,
				Output:                  getSymbol(method.GetOutputType()),
			}

			twirpSvc.Methods = append(twirpSvc.Methods, twirpMethod)
		}
		vars.Services = append(vars.Services, twirpSvc)
	}

	var buf = &bytes.Buffer{}
	err := TwirpTemplate.Execute(buf, vars)
	if err != nil {
		return nil, err
	}

	resp := &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(strings.TrimSuffix(name, path.Ext(name)) + "_twirp.py"),
		Content: proto.String(buf.String()),
	}

	return resp, nil
}

func getSymbol(name string) string {
	return strings.TrimPrefix(name, ".")
}

func getFileDescriptor(files []*descriptor.FileDescriptorProto, name string) (*descriptor.FileDescriptorProto, error) {
	//Assumption: Number of files will not be large enough to justify making a map
	for _, f := range files {
		if f.GetName() == name {
			return f, nil
		}
	}
	return nil, errors.New("could not find descriptor")
}
