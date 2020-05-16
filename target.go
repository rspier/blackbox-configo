package blackbox

/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

import (
	"bytes"
	"text/template"

	"github.com/golang/glog"
)

type Target struct {
	Module      string
	Destination string
	Name        string
}

type Targets struct {
	Targets          []Target
	JobName          string
	BlackboxHostPort string
	ScrapeInterval   int
}

func (ts *Targets) Add(m *Module, d, n string) {
	ts.Targets = append(ts.Targets, Target{m.Name, d, n})
}

var cfgTmpl = `
global:
  scrape_interval:     15s 
  evaluation_interval: 15s 

scrape_configs:
` + scCfgTmpl

var scCfgTmpl = `- job_name: '{{ .JobName }}'
  scrape_interval: {{ .ScrapeInterval }}s
  metrics_path: /probe
  static_configs:
  - targets: {{ range .Targets }}
    - {{.Module}}|{{.Destination}}|{{.Name}}{{end}}
  relabel_configs:
  - source_labels: [__address__]
    regex: (.+)\|(.+)\|(.+)
    target_label: __param_target
    replacement: ${2}
  - source_labels: [__address__]
    regex: (.+)\|(.+)\|(.+)
    target_label: __param_module
    replacement: ${1}
  - source_labels: [__address__]
    regex: (.+)\|(.+)\|(.+)
    target_label: module
    replacement: ${1}
  - source_labels: [__address__]
    regex: (.+)\|(.+)\|(.+)
    target_label: name
    replacement: ${3}
  - source_labels: [__param_target]
    target_label: instance
  - target_label: __address__
    replacement: {{.BlackboxHostPort}}
`

func (ts *Targets) Marshal() []byte {
	return ts.marshal(cfgTmpl)
}

func (ts *Targets) MarshalSC() []byte {
	return ts.marshal(scCfgTmpl)
}

func (ts *Targets) marshal(t string) []byte {
	tmpl, err := template.New("targets").Parse(t)
	if err != nil {
		glog.Fatal(err)
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, ts)
	if err != nil {
		glog.Fatal(err)
	}
	return b.Bytes()

}
