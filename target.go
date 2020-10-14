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
	Module         string
	Destination    string
	Name           string
	ScrapeInterval int
}

type Targets struct {
	Targets          []Target
	JobName          string
	BlackboxHostPort string
	ScrapeInterval   int
}

type TargetOption func(t *Target)

func (ts *Targets) Add(m *Module, d, n string, os ...*Option) Target {
	t := Target{
		Module:      m.Name,
		Destination: d,
		Name:        n,
	}
	t.applyOptions(os...)
	ts.Targets = append(ts.Targets, t)
	return t
}

func (t *Target) applyOptions(os ...*Option) {
	for _, o := range os {
		if o.TargetOption == nil {
			continue
		}
		o.TargetOption(t)
	}
}

func ScrapeInterval(si int) *Option {
	return &Option{
		TargetOption: func(t *Target) {
			t.ScrapeInterval = si
		},
	}
}

var header = `
global:
  scrape_interval:     15s 
  evaluation_interval: 15s 

scrape_configs:
`

var scCfgTmpl = `- job_name: '{{ .JobName }}_{{ .ScrapeInterval }}'
  scrape_interval: {{ .ScrapeInterval }}s
  metrics_path: /probe
  static_configs:
  - targets:{{ range .Targets }}
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
	var b bytes.Buffer
	b.WriteString(header)
	b.Write(ts.marshal())
	return b.Bytes()
}

func (ts *Targets) MarshalSC() []byte {
	return ts.marshal()
}

func (ts *Targets) byScrapeInterval() map[int][]Target {
	out := make(map[int][]Target)
	for _, t := range ts.Targets {
		si := t.ScrapeInterval
		if si == 0 {
			si = ts.ScrapeInterval
		}
		out[si] = append(out[si], t)
	}
	return out
}

var tmpl = template.Must(template.New("targets").Parse(scCfgTmpl))

func (ts *Targets) marshal() []byte {
	tsis := ts.byScrapeInterval()

	var b bytes.Buffer

	type tmplD struct {
		JobName          string
		ScrapeInterval   int
		Targets          []Target
		BlackboxHostPort string
	}

	for si, tsi := range tsis {
		d := &tmplD{
			JobName:          ts.JobName,
			BlackboxHostPort: ts.BlackboxHostPort,
			ScrapeInterval:   si,
			Targets:          tsi,
		}

		err := tmpl.Execute(&b, d)
		if err != nil {
			glog.Fatal(err)
		}
	}
	return b.Bytes()

}
