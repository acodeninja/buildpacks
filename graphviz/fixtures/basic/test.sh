#!/usr/bin/env bash

dot -Tsvg test.dot -O

diff test.dot.golden.svg test.dot.svg
