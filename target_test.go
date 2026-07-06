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
	"os"
	"path/filepath"
	"testing"
)

func TestTrimScheme(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://example.com",
			expected: "example.com",
		},
		{
			input:    "https://example.com",
			expected: "example.com",
		},
		{
			input:    "example.com",
			expected: "example.com",
		},
		{
			input:    "http://https://example.com",
			expected: "https://example.com",
		},
		{
			input:    "https://http://example.com",
			expected: "example.com",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "ftp://example.com",
			expected: "ftp://example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := trimScheme(tc.input)
			if got != tc.expected {
				t.Errorf("trimScheme(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestTargetsMarshalGolden(t *testing.T) {
	ts := &Targets{
		JobName:          "test_job",
		BlackboxHostPort: "localhost:9115",
		ScrapeInterval:   30,
	}

	m1 := &Module{Name: "http_200", Module: BaseHTTPModule(200)}
	m2 := &Module{Name: "redir_to_https_example_com", Module: BaseHTTPModule(302)}

	ts.Add(m1, "https://example.com", "example.com")
	ts.Add(m2, "http://example.com", "redir_to_https_example_com")
	ts.Add(m1, "https://slow.example.com", "slow.example.com", ScrapeInterval(60))

	got := ts.Marshal()

	goldenPath := filepath.Join("testdata", "targets_marshal.golden")

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
		t.Errorf("Targets.Marshal output mismatch\ngot:\n%s\nwant:\n%s", string(got), string(want))
	}
}

