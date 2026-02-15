#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="50-ci-n-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
This is a test for the -n flag.
Assigning a symbolic name on initial checkin.
EOF

# Execute ci with -nTAG
ci -q -i -u -m"initial checkin" -nTAG -wtester -d'2021-01-01 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci with -n flag (assign symbolic name)

-- options.conf --
{"args": ["-q","-i","-u","-m","initial checkin","-nTAG","-wtester","-d","2021-01-01 00:00:00Z","input.txt"] }

-- input.txt --
This is a test for the -n flag.
Assigning a symbolic name on initial checkin.

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
