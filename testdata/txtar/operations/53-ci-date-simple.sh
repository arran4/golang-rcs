#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT


OUT="53-ci-date-simple.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
A clock unwinds its shadow into rain,
While silver echoes fold a paper sky;
EOF

ci -q -i -u -m"r1" -wtester -d"2022-02-02 12:00:00-08:00" file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci checkin with explicit date offset

-- options.conf --
{"args": ["-q","-i","-u","-m","r1","-wtester","-d","2022-02-02 12:00:00-08:00","input.txt"] }

-- input.txt --
A clock unwinds its shadow into rain,
While silver echoes fold a paper sky;

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
