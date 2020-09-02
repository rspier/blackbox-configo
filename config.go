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
	"fmt"
	"regexp"
	"strings"

	bbconfig "github.com/prometheus/blackbox_exporter/config"
)

// Config represents a prometheus blackbox monitoring config, encompassing
// blackbox_exporter module configuration and prometheus target configuration.
type Config struct {
	Modules ModuleMap
	Targets *Targets
}

// We need to check both the production site on the CDN and the local version.

func (c *Config) AddSimpleRule(url string, os ...ModuleOption) {
	m := &Module{Name: "http_200", Module: BaseHTTPModule(200)}
	applyOptions(m, os...)
	c.Modules.Add(m)
	n := url
	if len(os) > 0 {
		n = m.Name
	}
	c.Targets.Add(m, url, n)
}

func (c *Config) AddSimpleRuleWithRedirect(url string, os ...ModuleOption) {
	c.AddSimpleRule(url, os...)
	if strings.HasPrefix(url, "https://") {
		c.AddHTTPSRedirRule(url, Status(301, 302, 308))
	}
}

var idChars = regexp.MustCompile(`[^A-Za-z0-9_]+`)

func cleanName(s string) string {
	s = strings.TrimSuffix(s, "/")
	return idChars.ReplaceAllString(s, "_")
}

func (c *Config) AddHTTPSRedirRule(in string, os ...ModuleOption) {
	src := strings.Replace(in, "https://", "http://", 1)
	dst := strings.Replace(in, "http://", "https://", 1)

	n := cleanName("redir_to_" + strings.TrimPrefix(dst, "http://"))

	os = append(os, Name(n))
	c.AddRedirRule(src, dst, os...)
}

func (c *Config) AddRedirRule(src, dst string, os ...ModuleOption) {
	m := RedirModule(302, dst)

	n := cleanName("redir_to_" + strings.TrimPrefix(dst, "http://"))
	os = append(os, Name(n))
	applyOptions(m, os...)

	c.Modules.Add(m)
	c.Targets.Add(m, src, m.Name)
}

func (c *Config) AddDNSRule(server, qtype, qname string, os ...ModuleOption) {
	m := DNSModule(qtype, qname)
	n := cleanName(fmt.Sprintf("dns_%s_%s", qname, qtype))
	os = append(os, Name(n))

	applyOptions(m, os...)

	c.Modules.Add(m)
	c.Targets.Add(m, server, m.Name)
}

func (c *Config) BBModules() bbconfig.Config {
	var bbm = make(map[string]bbconfig.Module)
	for n, m := range c.Modules {
		bbm[n] = *m.Module
	}

	bbc := bbconfig.Config{
		Modules: bbm,
	}

	return bbc
}
