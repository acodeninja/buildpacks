#!/usr/bin/env bash

PSQL_WHICH_TEST_OUTCOME=""
PSQL_WHICH="$(which psql)"
if [[ "$PSQL_WHICH" == "/layers/acodeninja_psql/psql/psql-bin/psql" ]]
then
  PSQL_WHICH_TEST_OUTCOME="passed 🟢"
else
  PSQL_WHICH_TEST_OUTCOME="psql is located at $PSQL_WHICH 🔴"
fi
echo "psql command location: $PSQL_WHICH_TEST_OUTCOME"

PSQL_VERSION_TEST_OUTCOME=""
PSQL_VERSION="$(psql --version)"
if [[ "$PSQL_VERSION" =~ "(PostgreSQL) 14" ]]
then
  PSQL_VERSION_TEST_OUTCOME="passed 🟢"
else
  PSQL_VERSION_TEST_OUTCOME="psql version is $PSQL_VERSION 🔴"
fi
echo "psql command version: $PSQL_VERSION_TEST_OUTCOME"

if [[ "$PSQL_VERSION_TEST_OUTCOME" == "passed 🟢" && "$PSQL_VERSION_TEST_OUTCOME" == "passed 🟢" ]]
then
  exit 0
else
  exit 1
fi
