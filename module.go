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
	"crypto/sha1"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	bbconfig "github.com/prometheus/blackbox_exporter/config"
	"gopkg.in/yaml.v3"
)

type Module struct {
	Name        string
	Description string
	Module      *bbconfig.Module
	HasOptions  bool
}

type ModuleMap map[string]*Module

func BaseHTTPModule(status int) *bbconfig.Module {
	c := &bbconfig.Module{
		Prober: "http",
		HTTP: bbconfig.HTTPProbe{
			ValidStatusCodes: []int{status},
			IPProtocol:       "ip4", // simpler to not worry about IPv6
		},
	}
	return c
}

func HTTPModule(status int) *Module {
	m := &Module{
		Name:   fmt.Sprintf("http_%d", status),
		Module: BaseHTTPModule(status),
	}
	return m
}

func boolPtr(b bool) *bool {
	return &b
}

// TODO: this doesn't actually use mm
func RedirModule(status int, dest string) *Module {
	bbm := BaseHTTPModule(status)

	bbm.HTTP.NoFollowRedirects = boolPtr(true)
	bbm.HTTP.FailIfHeaderNotMatchesRegexp = []bbconfig.HeaderMatch{{
		Header: "Location",
		Regexp: bbconfig.MustNewRegexp(quoteMeta(dest)),
	}}

	m := &Module{
		Description: fmt.Sprintf("%d to %v", status, dest),
		Module:      bbm,
	}
	return m
}

func BaseDNSModule() *bbconfig.Module {
	c := &bbconfig.Module{
		Prober: "dns",
		DNS: bbconfig.DNSProbe{
			IPProtocol: "ip4", // simpler to not worry about IPv6
		},
	}
	return c
}

func DNSModule(qtype, qname string) *Module {
	bbm := BaseDNSModule()
	bbm.DNS.QueryName = qname
	bbm.DNS.QueryType = qtype

	m := &Module{
		Description: fmt.Sprintf("dns query for %q", qname),
		Module:      bbm,
	}
	return m
}

func BaseTCPModule() *bbconfig.Module {
	c := &bbconfig.Module{
		Prober: "tcp",
		TCP: bbconfig.TCPProbe{
			IPProtocol: "ip4", // simpler to not worry about IPv6
		},
	}
	return c

}

func formatQueryResponse(qr []bbconfig.QueryResponse) string {
	var out string
	for _, q := range qr {
		out = out + fmt.Sprintf("%q -> %q,", q.Send, q.Expect)
	}
	return out

}

func TCPModule(qr []bbconfig.QueryResponse) *Module {
	tm := BaseTCPModule()
	tm.TCP.QueryResponse = qr

	m := &Module{
		Description: fmt.Sprintf("tcp module %s", formatQueryResponse(qr)),
		Module:      tm,
	}
	return m
}

func (m Module) hash() string {
	y, err := yaml.Marshal(m)
	if err != nil {
		glog.Fatalf("can't Marshal Module: %v", err)
	}
	h := sha1.Sum(y)
	return fmt.Sprintf("%x", h[0:4])
}

var seq = 0

func (mm ModuleMap) Add(m *Module) {
	// If this Module isn't already named, give it one.
	if len(m.Name) == 0 {
		m.Name = "mod_" + m.hash()
	}

	if existing, ok := mm[m.Name]; ok && m.HasOptions && existing.hash() != m.hash() {
		seq++
		m.Name += fmt.Sprintf("-%d", seq)
	}

	// Save the Module.  It's possible this overwrites an existing one, but if
	// the Name is the same, it should be the same
	mm[m.Name] = m
}

func quoteMeta(s string) string {
	return `\Q` + s + `\E`
}

type ModuleOption func(*Module)

func Status(s ...int) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Description += fmt.Sprintf("Status(%v) ", s)
			m.Module.HTTP.ValidStatusCodes = s
		},
	}
}

func Name(n string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			n = strings.ReplaceAll(n, " ", "-")
			m.Name = n
		}}
}

func Contains(cs ...string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Description += fmt.Sprintf("Contains(%v) ", cs)

			for _, c := range cs {
				m.Module.HTTP.FailIfBodyNotMatchesRegexp =
					append(m.Module.HTTP.FailIfBodyNotMatchesRegexp, bbconfig.MustNewRegexp(quoteMeta(c)))
			}
		}}

}

func NoFollowRedirects() *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Description += "NoFollowRedirects() "
			m.Module.HTTP.NoFollowRedirects = boolPtr(true)
		},
	}
}

func Header(h, v string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			if m.Module.HTTP.Headers == nil {
				m.Module.HTTP.Headers = make(map[string]string)
			}
			m.Module.HTTP.Headers[h] = v
		},
	}
}

func DNSAnswerFailIfMatchesRegexp(ms ...string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Module.DNS.ValidateAnswer.FailIfMatchesRegexp = ms
		},
	}
}

func DNSAnswerFailIfNotMatchesRegexp(ms ...string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Module.DNS.ValidateAnswer.FailIfNotMatchesRegexp = ms
		},
	}
}

func DNSAuthorityFailIfMatchesRegexp(ms ...string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Module.DNS.ValidateAuthority.FailIfMatchesRegexp = ms
		},
	}
}

func DNSAuthorityFailIfNotMatchesRegexp(ms ...string) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Module.DNS.ValidateAuthority.FailIfNotMatchesRegexp = ms
		},
	}
}

func TCPUseTLS() *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Module.TCP.TLS = true
			m.Name = m.Name + "_tls"
		},
	}
}

func Timeout(t time.Duration) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			m.Module.Timeout = t
		},
	}
}

func CustomFunc(f func(*bbconfig.Module)) *Option {
	return &Option{
		ModuleOption: func(m *Module) {
			f(m.Module)
		},
	}
}

func (m *Module) applyOptions(os ...*Option) {
	var once sync.Once
	for _, o := range os {
		if o.ModuleOption == nil {
			continue
		}
		once.Do(func() {
			if len(os) > 0 {
				m.Name = ""
				m.HasOptions = true
			}
		})
		o.ModuleOption(m)
	}
}
