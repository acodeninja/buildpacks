#!/usr/bin/env bash

TESTS_FAILED=""
CURRENT_TEST_GROUP_NAME=""

function test_group() {
  TEST_GROUP_NAME=$1
  CURRENT_TEST_GROUP_NAME="$TEST_GROUP_NAME"
}

function run_test() {
  TEST_NAME=$1
  TEST_COMMAND=$2
  TEST_EXPECTED_OUTPUT=$3

  TEST_OUTCOME=""
  TEST_OUTPUT=`$TEST_COMMAND 2>&1`
  if [[ "$TEST_OUTPUT" =~ "$TEST_EXPECTED_OUTPUT" ]]
  then
    TEST_OUTCOME="passed 🟢"
  else
    TEST_OUTCOME="failed 🔴 - $TEST_OUTPUT"
    TESTS_FAILED="yes"
  fi
  echo "[$CURRENT_TEST_GROUP_NAME][$TEST_NAME]: $TEST_OUTCOME"
}

COMMANDS_TO_TEST="pg_amcheck pgbench pg_config pg_dump pg_dumpall pg_isready pg_receivewal pg_restore psql"

for COMMAND_TO_TEST in $COMMANDS_TO_TEST; do
  test_group "$COMMAND_TO_TEST"
  run_test "location" "which $COMMAND_TO_TEST" "/layers/acodeninja_psql/psql/psql-bin/$COMMAND_TO_TEST"
  run_test "$COMMAND_TO_TEST version" "$COMMAND_TO_TEST --version" "14"
  run_test "$COMMAND_TO_TEST server" "$COMMAND_TO_TEST --version" "PostgreSQL"
done

QUERY_OUTPUT="$(psql "$DATABASE_URL" -c "SELECT datname FROM pg_database;")"

if [[ "$QUERY_OUTPUT" =~ "template1" ]]; then
  echo "[psql][run query]: passed 🟢"
else
  echo "[psql][run query]: failed 🔴"
  diff  <(echo "$QUERY_OUTPUT" ) <(echo "$QUERY_EXPECTED")
  TESTS_FAILED="yes"
fi

if [[ -z "$TESTS_FAILED" ]]
then
  exit 0
else
  exit 1
fi
