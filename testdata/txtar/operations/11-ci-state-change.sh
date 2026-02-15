#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

# Add RCS to PATH
export PATH=$HOME/rcs_install/bin:$PATH

OUT="11-ci-state-change.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
This is a test file for state change.
EOF

# Initial checkin
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

# Capture initial state as input.txt,v
cp file.txt,v input.txt,v

# Change state using rcs command
rcs -sProd file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs change state to Prod

-- options.conf --
{"args": ["-sProd","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
