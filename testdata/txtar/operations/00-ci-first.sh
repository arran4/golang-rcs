# 00-ci-first.sh
#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="00-ci-first.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
A clock unwinds its shadow into rain,
While silver echoes fold a paper sky;
EOF

ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci initial checkin (1.1)

-- options.conf --
{"args": ["-q","-i","-u","-m","r1","-wtester","-d","2020-01-01 00:00:00Z","input.txt"] }

-- input.txt --
A clock unwinds its shadow into rain,
While silver echoes fold a paper sky;

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
