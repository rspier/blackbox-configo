# Blackbox ConfigGo

A tool to generate [blackbox exporter](https://github.com/prometheus/blackbox_exporter)
configurations from Go code.  (No hand written YAML!)

## Why?

I was converting some configs from a very old Nagios, setup, and hand generating
the blackbox_exporter and prometheus configs was fragile and produced hard to
read configs.  So.... I wrote a tool.  Go provided free syntax checking,
formatting, and the joy of knowing that if it compiles, it'll probably run.

## Usage

See [the example](cmd/example/example.go) for what a config looks like.

The [Makefile](Makefile) is also useful.


Here's how you might want to use it to generate the files to integrate with an existing Prometheus server.

```bash
go run ./cmd/example \
    --scrape_interval=5m --blackbox=blackbox:9115 \
	--blackboxfile="blackbox-generated.yaml" \
	--targetsfile="scrape-config-generated.yaml" \
	--onlysc \
	--jobname="blackbox-generated"
```

## Tips

Use this tool to generate part of your file.  Have some static bits, and then
merge the output with [yq](https://github.com/mikefarah/yq).

Here's what a Makefile to do that might look like:

```make
scrape-config.yaml: scrape-config-base.yaml scrape-config-generated.yaml
        $(YQ) merge -a scrape-config-base.yaml scrape-config-generated.yaml > scrape-config.yaml

blackbox.yaml: blackbox-base.yaml blackbox-generated.yaml
        $(YQ) merge -a blackbox-base.yaml blackbox-generated.yaml > blackbox.yaml
```