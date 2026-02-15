#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="52-ci-k-preserve-metadata.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

cat > input.txt <<'EOF'
$Id: input.txt,v 1.1 1990/01/01 12:00:00 oldguy Exp $
Old content.
EOF

ci -q -k -u -m"restored" -t- input.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci -k preserves old metadata from keywords

-- options.conf --
{"args": ["-q","-k","-u","-m","restored","-t-","input.txt"] }

-- input.txt --
\$Id: input.txt,v 1.1 1990/01/01 12:00:00 oldguy Exp \$
Old content.

-- tests.txt --
ci

-- expected.txt,v --
$(cat input.txt,v)
EOF
