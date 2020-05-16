# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

default:
	echo Select a target.

# Example of running this for use in the real world with custom flags.
run-prod:
	go run ./cmd/example --scrape_interval=5m --blackbox=blackbox:9115 \
	  --blackboxfile="blackbox-generated.yaml" \
	  --targetsfile="scrape-config-generated.yaml" \
	  --onlysc \
	  --jobname="blackbox-generated"

# Just run it to generate the configs for testing.  Can be used with the test
# prometheus/blackbox below, because it also generates a prometheus.cfg for you.
run:
	go run ./cmd/example

# Run a local prometheus on port 9990 to test the config.
prometheus:
	docker run --rm --net host -v $(PWD):/cfg prom/prometheus \
	  --web.enable-lifecycle --config.file /cfg/prometheus.yaml --web.listen-address="0.0.0.0:9990"

# Run a local blackbox on port 9998 to test the config.
blackbox:
	docker run --rm --net host -v $(PWD):/cfg prom/blackbox-exporter \
	  --config.file /cfg/blackbox.yaml --web.listen-address="0.0.0.0:9998" --log.level=debug

# Reload the local prometheus and blackbox for testing.
reload:
	curl -X POST http://localhost:9990/-/reload
	curl -X POST http://localhost:9998/-/reload