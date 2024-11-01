#!/usr/bin/env bash

PERL5LIB="{{.PerlLibLocation}}"
PATH="{{.PostgresClientPath}}:$PATH"
LD_LIBRARY_PATH="{{range $val := .LibLocations}}{{$val}}:{{end}}$LD_LIBRARY_PATH"
LIBRARY_PATH="{{range $val := .LibLocations}}{{$val}}:{{end}}$LD_LIBRARY_PATH"

psql $@
