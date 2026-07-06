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
	"testing"
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
