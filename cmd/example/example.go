// An example of using blackbox-configo to generate a simple config.
package main

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
	bbconfig "github.com/prometheus/blackbox_exporter/config"
	bb "github.com/rspier/blackbox-configo"
)

func main() {
	bb.Main(config)
}

func config(c *bb.Config) {
	c.AddSimpleRuleWithRedirect(
		"https://www.cnn.com",
		bb.Contains("Breaking News"),
		bb.Name("cnn"))

	c.AddRedirRule("https://google.us/",
		"https://www.google.com/",
		bb.Status(302),
	)

	c.AddHTTPSRedirRule("http://golang.org/")

	c.AddDNSRule("8.8.8.8", "A", "www.firebase.com", bb.DNSAnswerFailIfNotMatchesRegexp("151.101.1.195", "151.101.65.195"))

	c.AddSMTPRule("localhost:25")
	c.AddIMAPRule("localhost:993", bb.TCPUseTLS(),
		bb.CustomFunc(func(m *bbconfig.Module) {
			m.TCP.TLSConfig.InsecureSkipVerify = true
		}))

}
