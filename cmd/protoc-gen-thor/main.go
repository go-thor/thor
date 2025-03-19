package main

import (
	"bytes"
	"log"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// Templates for code generation
var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-thor. DO NOT EDIT.
// source: {{ .FileName }}

package {{ .Package }}

import (
	"context"

	"github.com/go-thor/thor"
)
`))

	clientTemplate = template.Must(template.New("client").Parse(`
// {{ .ServiceName }}Client is the client API for {{ .ServiceName }} service.
type {{ .ServiceName }}Client struct {
	client pkg.Client
}

// New{{ .ServiceName }}Client creates a new client for the {{ .ServiceName }} service.
func New{{ .ServiceName }}Client(client pkg.Client) *{{ .ServiceName }}Client {
	return &{{ .ServiceName }}Client{client: client}
}
{{ range .Methods }}
// {{ .Name }} calls the {{ .Name }} method on the server.
func (c *{{ $.ServiceName }}Client) {{ .Name }}(ctx context.Context, in *{{ .InputType }}) (*{{ .OutputType }}, error) {
	out := new({{ .OutputType }})
	err := c.client.Call(ctx, "{{ $.ServiceName }}.{{ .Name }}", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// {{ .Name }}WithMetadata calls the {{ .Name }} method on the server with metadata.
func (c *{{ $.ServiceName }}Client) {{ .Name }}WithMetadata(ctx context.Context, in *{{ .InputType }}, metadata map[string]string) (*{{ .OutputType }}, error) {
	out := new({{ .OutputType }})
	err := c.client.CallWithMetadata(ctx, "{{ $.ServiceName }}.{{ .Name }}", in, out, metadata)
	if err != nil {
		return nil, err
	}
	return out, nil
}
{{ end }}
`))

	serverTemplate = template.Must(template.New("server").Parse(`
// {{ .ServiceName }}Server is the server API for {{ .ServiceName }} service.
type {{ .ServiceName }}Server interface {
{{ range .Methods }}	{{ .Name }}(ctx context.Context, req *{{ .InputType }}) (*{{ .OutputType }}, error)
{{ end }}}

// Register{{ .ServiceName }}Server registers the {{ .ServiceName }} server with pkg.Server.
func Register{{ .ServiceName }}Server(s pkg.Server, srv {{ .ServiceName }}Server) error {
	return s.RegisterName("{{ .ServiceName }}", &{{ .ServiceName }}ServerImpl{srv: srv})
}

// {{ .ServiceName }}ServerImpl implements the {{ .ServiceName }} service.
type {{ .ServiceName }}ServerImpl struct {
	srv {{ .ServiceName }}Server
}
{{ range .Methods }}
// {{ .Name }} implements {{ $.ServiceName }}Server.{{ .Name }}
func (s *{{ $.ServiceName }}ServerImpl) {{ .Name }}(ctx context.Context, req *{{ .InputType }}) (*{{ .OutputType }}, error) {
	return s.srv.{{ .Name }}(ctx, req)
}
{{ end }}
`))
)

// TemplateData holds data for template rendering
type TemplateData struct {
	FileName    string
	Package     string
	ServiceName string
	Methods     []MethodData
}

// MethodData holds method information for template rendering
type MethodData struct {
	Name       string
	InputType  string
	OutputType string
}

func main() {
	// Run the plugin with default options
	protogen.Options{
		ParamFunc: func(name, value string) error {
			return nil
		},
	}.Run(generatePlugin)
}

// generatePlugin implements the plugin logic
func generatePlugin(gen *protogen.Plugin) error {
	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}
		generateFile(gen, file)
	}
	return nil
}

// generateFile generates code for a single .proto file.
func generateFile(gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}

	// Determine output filename
	fileName := file.GeneratedFilenamePrefix + ".thor.go"
	gfile := gen.NewGeneratedFile(fileName, file.GoImportPath)

	// Create template data
	data := TemplateData{
		FileName: file.Desc.Path(),
		Package:  string(file.GoPackageName),
	}

	// Generate header
	var buf bytes.Buffer
	if err := headerTemplate.Execute(&buf, data); err != nil {
		log.Fatalf("Failed to execute header template: %v", err)
	}

	// Generate code for each service
	for _, service := range file.Services {
		serviceData := TemplateData{
			FileName:    file.Desc.Path(),
			Package:     string(file.GoPackageName),
			ServiceName: service.GoName,
			Methods:     make([]MethodData, 0, len(service.Methods)),
		}

		// Add method information
		for _, method := range service.Methods {
			methodData := MethodData{
				Name:       method.GoName,
				InputType:  method.Input.GoIdent.GoName,
				OutputType: method.Output.GoIdent.GoName,
			}
			serviceData.Methods = append(serviceData.Methods, methodData)
		}

		// Generate client code
		if err := clientTemplate.Execute(&buf, serviceData); err != nil {
			log.Fatalf("Failed to execute client template: %v", err)
		}

		// Generate server code
		if err := serverTemplate.Execute(&buf, serviceData); err != nil {
			log.Fatalf("Failed to execute server template: %v", err)
		}
	}

	// Write generated code to file
	gfile.Write(buf.Bytes())
}
