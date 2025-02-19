#!/usr/bin/env bash

TESTS_FAILED=""

if [[ -f .test-created-file ]]; then
  echo "[file-created]: passed 🟢"
else
  echo "[file-created]: failed 🔴"
  TESTS_FAILED="yes"
fi

if [[ -f .pytest-ran ]]; then
  echo "[pytest]: passed 🟢"
else
  echo "[pytest]: failed 🔴"
  TESTS_FAILED="yes"
fi

if [[ -z "$TESTS_FAILED" ]]
then
  exit 0
else
  exit 1
fi
