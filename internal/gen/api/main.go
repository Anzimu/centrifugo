package main

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/centrifugal/centrifugo/v5/internal/gen"
)

func main() {
	generateHandlersHTTP()
	generateHandlersGRPC()
	generateRequestDecoder()
	generateResponseEncoder()
	generateResultEncoder()
}

type TemplateData struct {
	RequestCapitalized string
	RequestLower       string
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func generateToFile(header, funcTmpl, outFile string, excludeRequests []string) {
	tmpl := template.Must(template.New("").Parse(funcTmpl))

	var buf bytes.Buffer
	buf.WriteString(header)

	for _, req := range gen.Requests {
		if stringInSlice(req, excludeRequests) {
			continue
		}
		err := tmpl.Execute(&buf, TemplateData{
			RequestCapitalized: req,
			RequestLower:       strings.ToLower(req),
		})
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()
	_, _ = buf.WriteTo(file)
}

var headerHandlersHTTP = `// Code generated by internal/gen/api/main.go. DO NOT EDIT.

package api

import (
	"io"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)
`

var templateFuncHandlersHTTP = `
func (s *Handler) handle{{ .RequestCapitalized }}(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		s.handleReadDataError(r, w, err)
		return
	}

	req, err := paramsDecoder.Decode{{ .RequestCapitalized }}(data)
	if err != nil {
		s.handleUnmarshalError(r, w, err)
		return
	}

	resp := s.api.{{ .RequestCapitalized }}(r.Context(), req)

{{- if ne .RequestCapitalized "Batch" }}
	if s.config.UseOpenTelemetry && resp.Error != nil {
		span := trace.SpanFromContext(r.Context())
		span.SetStatus(codes.Error, resp.Error.Error())
	}
{{- end}}

	data, err = responseEncoder.Encode{{ .RequestCapitalized }}(resp)
	if err != nil {
		s.handleMarshalError(r, w, err)
		return
	}

	s.writeJson(w, data)
}
`

func generateHandlersHTTP() {
	generateToFile(headerHandlersHTTP, templateFuncHandlersHTTP, "internal/api/handler_gen.go", nil)
}

var headerHandlersGRPC = `// Code generated by internal/gen/api/main.go. DO NOT EDIT.

package api

import (
	"context"

	. "github.com/centrifugal/centrifugo/v5/internal/apiproto"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)
`

var templateFuncHandlersGRPC = `
// {{ .RequestCapitalized }} ...
func (s *grpcAPIService) {{ .RequestCapitalized }}(ctx context.Context, req *{{ .RequestCapitalized }}Request) (*{{ .RequestCapitalized }}Response, error) {
	resp := s.api.{{ .RequestCapitalized }}(ctx, req)
{{- if ne .RequestCapitalized "Batch" }}
	if s.useOpenTelemetry && resp.Error != nil {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, resp.Error.Error())
	}
{{- end}}
	return resp, nil
}
`

func generateHandlersGRPC() {
	generateToFile(headerHandlersGRPC, templateFuncHandlersGRPC, "internal/api/grpc_handler_gen.go", nil)
}

var headerRequestDecoder = `// Code generated by internal/gen/api/main.go. DO NOT EDIT.

package apiproto

import "encoding/json"

var _ RequestDecoder = (*JSONRequestDecoder)(nil)

// JSONRequestDecoder ...
type JSONRequestDecoder struct{}

// NewJSONRequestDecoder ...
func NewJSONRequestDecoder() *JSONRequestDecoder {
	return &JSONRequestDecoder{}
}
`

var templateFuncRequestDecoder = `
// Decode{{ .RequestCapitalized }} ...
func (d *JSONRequestDecoder) Decode{{ .RequestCapitalized }}(data []byte) (*{{ .RequestCapitalized }}Request, error) {
	var p {{ .RequestCapitalized }}Request
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
`

func generateRequestDecoder() {
	generateToFile(headerRequestDecoder, templateFuncRequestDecoder, "internal/apiproto/decode_request_gen.go", nil)
}

var headerResponseEncoder = `// Code generated by internal/gen/api/main.go. DO NOT EDIT.

package apiproto

import "encoding/json"

// JSONResponseEncoder ...
type JSONResponseEncoder struct{}

func NewJSONResponseEncoder() *JSONResponseEncoder {
	return &JSONResponseEncoder{}
}
`

var templateFuncResponseEncoder = `
// Encode{{ .RequestCapitalized }} ...
func (e *JSONResponseEncoder) Encode{{ .RequestCapitalized }}(response *{{ .RequestCapitalized }}Response) ([]byte, error) {
	return json.Marshal(response)
}
`

func generateResponseEncoder() {
	generateToFile(headerResponseEncoder, templateFuncResponseEncoder, "internal/apiproto/encode_response_gen.go", nil)
}

var headerResultEncoder = `// Code generated by internal/gen/api/main.go. DO NOT EDIT.

package apiproto

import "encoding/json"

var _ ResultEncoder = (*JSONResultEncoder)(nil)

// JSONResultEncoder ...
type JSONResultEncoder struct{}

// NewJSONResultEncoder ...
func NewJSONResultEncoder() *JSONResultEncoder {
	return &JSONResultEncoder{}
}
`

var templateFuncResultEncoder = `
// Encode{{ .RequestCapitalized }} ...
func (e *JSONResultEncoder) Encode{{ .RequestCapitalized }}(res *{{ .RequestCapitalized }}Result) ([]byte, error) {
	return json.Marshal(res)
}
`

func generateResultEncoder() {
	generateToFile(headerResultEncoder, templateFuncResultEncoder, "internal/apiproto/encode_result_gen.go", []string{"Batch"})
}
