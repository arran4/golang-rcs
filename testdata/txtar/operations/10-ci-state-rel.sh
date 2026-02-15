#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

# Add RCS to PATH
export PATH=$HOME/rcs_install/bin:$PATH

OUT="10-ci-state-rel.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
This is a test file for state Rel.
EOF

ci -q -i -u -m"r1" -sRel -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci checkin with state Rel (1.1)

-- options.conf --
{"args": ["-q","-i","-u","-m","r1","-sRel","-wtester","-d","2020-01-01 00:00:00Z","input.txt"] }

-- input.txt --
This is a test file for state Rel.

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
