#!/usr/bin/env bash
# Test runner for the Expression Transformation Project.
# Walks tests/*.in, compares stdout to the matching .out, and separately
# walks tests/err/*.in expecting a non-zero exit code.

set -u

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
PY="${PYTHON:-python3}"
MAIN="$ROOT/expr.py"

pass=0
fail=0
failed_names=()

run_ok() {
    local in_file="$1"
    local out_file="${in_file%.in}.out"
    local name
    name="$(basename "${in_file%.in}")"
    local got
    got="$("$PY" "$MAIN" < "$in_file")"
    local rc=$?
    if [ $rc -ne 0 ]; then
        fail=$((fail+1)); failed_names+=("$name (expected success, exit=$rc)"); return
    fi
    if [ ! -f "$out_file" ]; then
        fail=$((fail+1)); failed_names+=("$name (missing .out)"); return
    fi
    local want
    want="$(cat "$out_file")"
    if [ "$got" = "$want" ]; then
        pass=$((pass+1))
    else
        fail=$((fail+1))
        failed_names+=("$name")
        echo "---- $name ----"
        echo "  want: $want"
        echo "  got : $got"
    fi
}

run_err() {
    local in_file="$1"
    local name
    name="$(basename "${in_file%.in}")"
    "$PY" "$MAIN" < "$in_file" >/dev/null 2>&1
    local rc=$?
    if [ $rc -ne 0 ]; then
        pass=$((pass+1))
    else
        fail=$((fail+1)); failed_names+=("err/$name (expected failure, exit=0)")
    fi
}

shopt -s nullglob
for f in "$HERE"/*.in;        do run_ok  "$f"; done
for f in "$HERE"/err/*.in;    do run_err "$f"; done

echo "---------------------------------------"
echo "Passed: $pass    Failed: $fail"
if [ $fail -ne 0 ]; then
    printf '  - %s\n' "${failed_names[@]}"
    exit 1
fi
