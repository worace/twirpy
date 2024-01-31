package generator

import "text/template"

type TwirpTemplateVariables struct {
	FileName       string
	Services       []*TwirpService
	SchemaFileName string
}

type TwirpService struct {
	ServiceURL string
	Name       string
	Comment    string
	Methods    []*TwirpMethod
}

type TwirpMethod struct {
	ServiceURL              string
	ServiceName             string
	Name                    string
	Comment                 string
	Input                   string
	InputMessageName        string
	InputMessageModuleName  string
	OutputMessageName       string
	OutputMessageModuleName string
	Output                  string
}

type TwirpImport struct {
	From   string
	Import string
}

// TwirpTemplate - Template for twirp server and client
// {{.Input}} = twirp.example.haberdasher.Size
// need: from haberdasher_pb2 import Size
var TwirpTemplate = template.Must(template.New("TwirpTemplate").Parse(`# -*- coding: utf-8 -*-
# Generated by https://github.com/verloop/twirpy/protoc-gen-twirpy.  DO NOT EDIT!
# source: {{.FileName}}

from google.protobuf import symbol_database as _symbol_database

from twirp.base import Endpoint
from twirp.server import TwirpServer
from twirp.client import TwirpClient
from twirp.context import Context
from abc import ABC, abstractmethod

{{- range .Services}}
{{- range .Methods}}
from {{.InputMessageModuleName}}_pb2 import {{.InputMessageName}}
from {{.OutputMessageModuleName}}_pb2 import {{.OutputMessageName}}
{{- end}}
{{- end}}




_sym_db = _symbol_database.Default()

{{range .Services}}

class {{.Name}}Service(ABC):
	def __init__(self, *args, **kwargs):
		super().__init__(*args, **kwargs)
		self.endpoints = { {{- range .Methods }}
			"{{.Name}}": Endpoint(
				service_name="{{.ServiceName}}",
				name="{{.Name}}",
				function=getattr(self, "{{.Name}}"),
				input=_sym_db.GetSymbol("{{.Input}}"),
				output=_sym_db.GetSymbol("{{.Output}}"),
			),{{- end }}
		}

	# package.ServiceName e.g. haberdasher.Haberdasher
	service_id: str = "{{.ServiceURL}}"
	{{range .Methods}}
	@abstractmethod
	def {{.Name}}(self, ctx: Context, arg: {{.InputMessageName}}) -> {{.OutputMessageName}}:
		raise NotImplementedError()
	{{end}}


class {{.Name}}Server(TwirpServer):

	def __init__(self, *args, service: {{.Name}}Service, server_path_prefix="/twirp"):
		super().__init__(service=service)
		self._prefix = F"{server_path_prefix}/{{.ServiceURL}}"
		self._endpoints = { {{- range .Methods }}
			"{{.Name}}": Endpoint(
				service_name="{{.ServiceName}}",
				name="{{.Name}}",
				function=getattr(service, "{{.Name}}"),
				input=_sym_db.GetSymbol("{{.Input}}"),
				output=_sym_db.GetSymbol("{{.Output}}"),
			),{{- end }}
		}

class {{.Name}}Client(TwirpClient):
{{range .Methods}}
	def {{.Name}}(self, *args, ctx: Context, request: {{.InputMessageName}}, server_path_prefix="/twirp", **kwargs) -> {{.OutputMessageName}}:
		return self._make_request(
			url=F"{server_path_prefix}/{{.ServiceURL}}/{{.Name}}",
			ctx=ctx,
			request=request,
			response_obj=_sym_db.GetSymbol("{{.Output}}"),
			**kwargs,
		)
{{end}}


# This should be moved to the core twirp runtime lib but for now i'm inlining it in the app code
# since i don't have a good way to publish a fork of the lib
from libs.proto_utils.twirp_runtime import LocalTwirpClient

class Local{{.Name}}Client(LocalTwirpClient):
{{range .Methods}}
	def {{.Name}}(self, request: {{.InputMessageName}}) -> {{.OutputMessageName}}:
		return self._make_request(
			# will be like haberdasher.Haberdasher/MakeHat
			url=F"{{.ServiceURL}}/{{.Name}}",
			request=request,
		)
{{end}} # end Range .Methods
{{end}} # end Range .Services
`))
