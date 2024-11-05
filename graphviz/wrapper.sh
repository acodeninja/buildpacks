#!/usr/bin/env bash

PATH="{{.Path}}:$PATH" \
LD_LIBRARY_PATH="{{range $val := .LibLocations}}{{$val}}:{{end}}$LD_LIBRARY_PATH" \
LIBRARY_PATH="{{range $val := .LibLocations}}{{$val}}:{{end}}$LD_LIBRARY_PATH"  \
FONTCONFIG_PATH="{{.FontConfigLocation}}" \
GVBINDIR="{{.GraphvizBinDir}}" \
{{.Command}} $@
