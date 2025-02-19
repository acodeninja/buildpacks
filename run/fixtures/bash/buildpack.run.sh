#!/usr/bin/env bash

set -ex

touch .test-created-file

python -m pytest .
