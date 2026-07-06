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
	"strings"
	"testing"
)

func TestQuoteMeta(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "example.com",
			expected: `\Qexample.com\E`,
		},
		{
			input:    "",
			expected: `\Q\E`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := quoteMeta(tc.input)
			if got != tc.expected {
				t.Errorf("quoteMeta(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestModuleMapAdd(t *testing.T) {
	mm := make(ModuleMap)

	// 1. Auto-naming when name is empty
	m1 := &Module{
		Module: BaseHTTPModule(200),
	}
	mm.Add(m1)
	if len(m1.Name) == 0 || !strings.HasPrefix(m1.Name, "mod_") {
		t.Errorf("expected module to be auto-named with 'mod_' prefix, got: %q", m1.Name)
	}

	// 2. Renaming collision resolution when names match, configurations differ, and HasOptions is true
	m2 := &Module{
		Name:       "duplicate_name",
		Module:     BaseHTTPModule(200),
		HasOptions: true,
	}
	m3 := &Module{
		Name:       "duplicate_name",
		Module:     BaseHTTPModule(404),
		HasOptions: true,
	}

	mm.Add(m2)
	mm.Add(m3)

	if m2.Name != "duplicate_name" {
		t.Errorf("expected original module name to remain unchanged, got %q", m2.Name)
	}
	if m3.Name == "duplicate_name" {
		t.Errorf("expected collision to rename second module, but got %q", m3.Name)
	}
	if !strings.HasPrefix(m3.Name, "duplicate_name-") {
		t.Errorf("expected collision name to have suffix like '-1', got %q", m3.Name)
	}
}

