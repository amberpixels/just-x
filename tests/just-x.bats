#!/usr/bin/env bats

JUST_X="$BATS_TEST_DIRNAME/../just-x"

@test "version prints version" {
  run "$JUST_X" version
  [ "$status" -eq 0 ]
  [[ "$output" == just-x* ]]
}

@test "help prints usage" {
  run "$JUST_X" help
  [ "$status" -eq 0 ]
  [[ "$output" == *"Extended recipe names"* ]]
}

@test "translate dev! → dev-x" {
  run "$JUST_X" translate "dev!"
  [ "$status" -eq 0 ]
  [[ "$output" == *"dev-x"* ]]
}

@test "translate ready? → ready-q" {
  run "$JUST_X" translate "ready?"
  [ "$status" -eq 0 ]
  [[ "$output" == *"ready-q"* ]]
}

@test "translate app:build → app--build" {
  run "$JUST_X" translate "app:build"
  [ "$status" -eq 0 ]
  [[ "$output" == *"app--build"* ]]
}

@test "translate with custom env vars" {
  JUST_X_BANG="-bang" run "$JUST_X" translate "dev!"
  [ "$status" -eq 0 ]
  [[ "$output" == *"dev-bang"* ]]
}

@test "translate plain recipe unchanged" {
  run "$JUST_X" translate "test"
  [ "$status" -eq 0 ]
  [[ "$output" == "test → test" ]]
}

@test "init without args defaults to just" {
  run "$JUST_X" init
  [ "$status" -eq 0 ]
  [[ "$output" == *"_just_x_just()"* ]]
  [[ "$output" == *"alias just="* ]]
  [[ "$output" == *"noglob"* ]]
  [[ "$output" == *"command just"* ]]
}

@test "init generates function with given name" {
  run "$JUST_X" init j
  [ "$status" -eq 0 ]
  [[ "$output" == *"_just_x_j()"* ]]
  [[ "$output" == *"alias j="* ]]
  [[ "$output" == *"noglob"* ]]
  [[ "$output" == *"command just"* ]]
}

@test "init generates function with custom name" {
  run "$JUST_X" init myalias
  [ "$status" -eq 0 ]
  [[ "$output" == *"_just_x_myalias()"* ]]
  [[ "$output" == *"alias myalias="* ]]
}

@test "unknown command fails" {
  run "$JUST_X" foobar
  [ "$status" -eq 1 ]
}

@test "generated function passes through plain commands" {
  eval "$("$JUST_X" init _test_j)"
  # We can't easily test the actual just call, but we can verify the function exists
  type _just_x__test_j | head -1
  [[ "$(type _just_x__test_j)" == *"function"* ]]
}

# --- Integration tests (require just to be installed) ---

setup_justfile() {
  TEST_DIR="$(mktemp -d)"
  cat > "$TEST_DIR/justfile" << 'JUSTFILE'
my_var := "default"

greet-x target arg1="def":
    @echo "var={{my_var}} target={{target}} arg1={{arg1}}"

plain target:
    @echo "target={{target}}"
JUSTFILE
  eval "$("$JUST_X" init _tj)"
}

teardown_justfile() {
  rm -rf "$TEST_DIR"
}

@test "integration: recipe with positional param" {
  setup_justfile
  run bash -c "cd '$TEST_DIR' && source <('$JUST_X' init _tj) && _just_x__tj 'greet!' mytarget"
  teardown_justfile
  [ "$status" -eq 0 ]
  [[ "$output" == *"var=default target=mytarget arg1=def"* ]]
}

@test "integration: recipe with all params" {
  setup_justfile
  run bash -c "cd '$TEST_DIR' && source <('$JUST_X' init _tj) && _just_x__tj 'greet!' mytarget custom_val"
  teardown_justfile
  [ "$status" -eq 0 ]
  [[ "$output" == *"var=default target=mytarget arg1=custom_val"* ]]
}

@test "integration: variable override + recipe with params" {
  setup_justfile
  run bash -c "cd '$TEST_DIR' && source <('$JUST_X' init _tj) && _just_x__tj my_var=hello 'greet!' mytarget"
  teardown_justfile
  [ "$status" -eq 0 ]
  [[ "$output" == *"var=hello target=mytarget arg1=def"* ]]
}

@test "integration: variable override + recipe + all params" {
  setup_justfile
  run bash -c "cd '$TEST_DIR' && source <('$JUST_X' init _tj) && _just_x__tj my_var=hello 'greet!' mytarget custom_val"
  teardown_justfile
  [ "$status" -eq 0 ]
  [[ "$output" == *"var=hello target=mytarget arg1=custom_val"* ]]
}

@test "integration: variable override + plain recipe + params" {
  setup_justfile
  run bash -c "cd '$TEST_DIR' && source <('$JUST_X' init _tj) && _just_x__tj my_var=hello plain mytarget"
  teardown_justfile
  [ "$status" -eq 0 ]
  [[ "$output" == *"target=mytarget"* ]]
}

@test "integration: recipe param that looks like VAR=VALUE is not consumed" {
  setup_justfile
  run bash -c "cd '$TEST_DIR' && source <('$JUST_X' init _tj) && _just_x__tj plain 'foo=bar'"
  teardown_justfile
  [ "$status" -eq 0 ]
  [[ "$output" == *"target=foo=bar"* ]]
}
