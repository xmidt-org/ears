#!/bin/bash

./otelcontribcol_darwin_amd64 --config otel_collector_config.yaml --log-level DEBUG --log-format json > otel.log 2>&1 &



