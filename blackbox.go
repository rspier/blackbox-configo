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
	"flag"
	"io/ioutil"
	"time"

	"github.com/golang/glog"
)

/*
	SimpleCheck -> https returns 200, http redirects to https
	first uses http_200 module, second uses custom module

	custom labels: "site" (www, rtperl, rtcpan)

    options: HTTPRedirCode(301 or 302)

	modules: x-site-http_301
*/

var (
	scrapeInterval = flag.Duration("scrape_interval", 30*time.Second, "scrape interval")
	blackbox       = flag.String("blackbox", "localhost:9998", "hostport of blackbox exporter")
	targetsFile    = flag.String("targetsfile", "prometheus.yaml", "file to write the generated targets to")
	blackboxFile   = flag.String("blackboxfile", "blackbox.yaml", "file to write the generated blackbox config to")
	onlySC         = flag.Bool("onlysc", false, "if true, only write out scrapeconfigs")
	jobName        = flag.String("jobname", "blackbox", "job_name for the target definition")
	groupName      = flag.String("group", "perl", "conifguration group")
)

// Main is the generic Main function.  Pass it a function that uses the Config object, and it will handle flags and output.
func Main(cfg func(c *Config)) {
	flag.Parse()

	c := &Config{
		Modules: make(ModuleMap),
		Targets: &Targets{
			BlackboxHostPort: *blackbox,
			ScrapeInterval:   int(scrapeInterval.Seconds()),
			JobName:          *jobName,
		},
	}

	cfg(c)

	cbs, err := c.Marshal()
	if err != nil {
		glog.Fatal(err)
	}
	ioutil.WriteFile(*blackboxFile, cbs, 0666)
	if !*onlySC {
		ioutil.WriteFile(*targetsFile, c.Targets.Marshal(), 0666)
	} else {
		ioutil.WriteFile(*targetsFile, c.Targets.MarshalSC(), 0666)
	}
}
