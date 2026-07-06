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
