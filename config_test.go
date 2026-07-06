package blackbox

/*
Copyright 2026 Robert Spier

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
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bbconfig "github.com/prometheus/blackbox_exporter/config"
)

func TestCleanName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "example.com/",
			expected: "example_com",
		},
		{
			input:    "foo/bar/",
			expected: "foo_bar",
		},
		{
			input:    "foo-bar",
			expected: "foo_bar",
		},
		{
			input:    "foo_bar",
			expected: "foo_bar",
		},
		{
			input:    "foo#bar$baz",
			expected: "foo_bar_baz",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "/",
			expected: "",
		},
		{
			input:    "abc/def/ghi",
			expected: "abc_def_ghi",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := cleanName(tc.input)
			if got != tc.expected {
				t.Errorf("cleanName(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

var update = flag.Bool("update", false, "update golden files")

func TestConfigMarshalGolden(t *testing.T) {
	c := &Config{
		Modules: make(ModuleMap),
		Targets: &Targets{},
	}
	c.Modules.Add(&Module{
		Name:   "http_200",
		Module: BaseHTTPModule(200),
	})
	c.Modules.Add(&Module{
		Name:   "http_404",
		Module: BaseHTTPModule(404),
	})

	got, err := c.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	goldenPath := filepath.Join("testdata", "config_marshal.golden")

	if *update {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, got, 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file (run with -update to generate): %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("Marshal output mismatch\ngot:\n%s\nwant:\n%s", string(got), string(want))
	}
}

func TestAddSimpleRuleWithRedirect(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		expectedTargets []Target
		expectedModules []string
	}{
		{
			name: "HTTP URL - no redirect",
			url:  "http://example.com",
			expectedTargets: []Target{
				{Module: "http_200", Destination: "http://example.com", Name: "http://example.com"},
			},
			expectedModules: []string{"http_200"},
		},
		{
			name: "HTTPS URL - with redirect",
			url:  "https://example.com",
			expectedTargets: []Target{
				{Module: "http_200", Destination: "https://example.com", Name: "https://example.com"},
				{Module: "redir_to_https_example_com", Destination: "http://example.com", Name: "redir_to_https_example_com"},
			},
			expectedModules: []string{"http_200", "redir_to_https_example_com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{
				Modules: make(ModuleMap),
				Targets: &Targets{},
			}

			c.AddSimpleRuleWithRedirect(tc.url)

			// Verify targets
			if len(c.Targets.Targets) != len(tc.expectedTargets) {
				t.Fatalf("expected %d targets, got %d", len(tc.expectedTargets), len(c.Targets.Targets))
			}
			for i, gotT := range c.Targets.Targets {
				wantT := tc.expectedTargets[i]
				if gotT.Module != wantT.Module || gotT.Destination != wantT.Destination || gotT.Name != wantT.Name {
					t.Errorf("target %d mismatch: got %+v, want %+v", i, gotT, wantT)
				}
			}

			// Verify modules exist
			for _, mName := range tc.expectedModules {
				if _, ok := c.Modules[mName]; !ok {
					t.Errorf("expected module %q to be added to Config.Modules", mName)
				}
			}
		})
	}
}

func TestMoreRulesAndOptions(t *testing.T) {
	c := &Config{
		Modules: make(ModuleMap),
		Targets: &Targets{},
	}

	// 1. AddDNSRule with options
	c.AddDNSRule("8.8.8.8", "A", "example.com",
		DNSAnswerFailIfMatchesRegexp("1.2.3.4"),
		DNSAnswerFailIfNotMatchesRegexp("5.6.7.8"),
		DNSAuthorityFailIfMatchesRegexp("9.9.9.9"),
		DNSAuthorityFailIfNotMatchesRegexp("0.0.0.0"),
		Timeout(2*time.Second),
	)

	// 2. AddTCPRule with options
	c.AddTCPRule("localhost:80",
		[]bbconfig.QueryResponse{{Send: "hello", Expect: bbconfig.MustNewRegexp("world")}},
		TCPUseTLS(),
		CustomFunc(func(m *bbconfig.Module) {
			m.TCP.QueryResponse[0].Send = "custom"
		}),
	)

	// 3. AddSMTPRule, AddIMAPRule, AddNNTPRule
	c.AddSMTPRule("mail.example.com")
	c.AddIMAPRule("imap.example.com")
	c.AddNNTPRule("nntp.example.com")

	// 4. AddSimpleRule with NoFollowRedirects, Header, Contains options
	c.AddSimpleRule("http://example.com/contains",
		NoFollowRedirects(),
		Header("X-Custom", "Value"),
		Contains("secret"),
	)

	// Verify modules count and some specific options
	bbm := c.BBModules()
	if len(bbm.Modules) != 6 {
		t.Errorf("expected 6 modules in BBModules, got %d", len(bbm.Modules))
	}

	// Verify DNS module properties
	dnsModName := cleanName("dns_example.com_A")
	dnsMod, ok := c.Modules[dnsModName]
	if !ok {
		t.Fatalf("expected DNS module %q to exist", dnsModName)
	}
	if dnsMod.Module.DNS.QueryName != "example.com" || dnsMod.Module.DNS.QueryType != "A" {
		t.Errorf("incorrect DNS query configuration: %+v", dnsMod.Module.DNS)
	}
	if dnsMod.Module.Timeout != 2*time.Second {
		t.Errorf("expected timeout of 2s, got %v", dnsMod.Module.Timeout)
	}

	// Verify custom function modified the TCP Send
	tcpMod, ok := c.Modules["mod_d2c2069e_tls"] // autonamed but with _tls suffix
	if !ok {
		// fallback search since name might vary by hash implementation/details
		for name, mod := range c.Modules {
			if strings.HasSuffix(name, "_tls") && mod.Module.Prober == "tcp" {
				tcpMod = mod
				ok = true
				break
			}
		}
	}
	if !ok {
		t.Fatalf("expected TCP module to exist")
	}
	if tcpMod.Module.TCP.QueryResponse[0].Send != "custom" {
		t.Errorf("expected custom function option to change Send value, got %q", tcpMod.Module.TCP.QueryResponse[0].Send)
	}
}

func TestHTTPModule(t *testing.T) {
	m := HTTPModule(201)
	if m.Name != "http_201" {
		t.Errorf("expected module name http_201, got %q", m.Name)
	}
	if m.Module.HTTP.ValidStatusCodes[0] != 201 {
		t.Errorf("expected status code 201, got %v", m.Module.HTTP.ValidStatusCodes)
	}
}


