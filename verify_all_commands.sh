
echo "Testing gorcs commands with no args (via script to simulate TTY) (should print usage)"
script -q -c "go run ./cmd/gorcs format" output_format.log
OUTPUT=$(cat output_format.log)
if echo "$OUTPUT" | grep -q "Usage:"; then
    echo "SUCCESS: Usage printed for format"
else
    echo "FAILURE: Usage NOT printed for format"
    echo "Output was:"
    cat output_format.log
fi

script -q -c "go run ./cmd/gorcs list-heads" output_list.log
OUTPUT=$(cat output_list.log)
if echo "$OUTPUT" | grep -q "Usage:"; then
    echo "SUCCESS: Usage printed for list-heads"
else
    echo "FAILURE: Usage NOT printed for list-heads"
    echo "Output was:"
    cat output_list.log
fi

script -q -c "go run ./cmd/gorcs normalize-revisions" output_norm.log
OUTPUT=$(cat output_norm.log)
if echo "$OUTPUT" | grep -q "Usage:"; then
    echo "SUCCESS: Usage printed for normalize-revisions"
else
    echo "FAILURE: Usage NOT printed for normalize-revisions"
    echo "Output was:"
    cat output_norm.log
fi

script -q -c "go run ./cmd/gorcs validate" output_val.log
OUTPUT=$(cat output_val.log)
if echo "$OUTPUT" | grep -q "Usage:"; then
    echo "SUCCESS: Usage printed for validate"
else
    echo "FAILURE: Usage NOT printed for validate"
    echo "Output was:"
    cat output_val.log
fi

echo "Testing gorcs format with piped input (no args) (should NOT print usage)"
OUTPUT=$(echo "head 1.1; access; symbols; locks; comment @# @; 1.1 date 2021.01.01.00.00.00; author u; state Exp; branches; next; desc @@ 1.1 log @@ text @@" | go run ./cmd/gorcs format -s 2>&1)
if echo "$OUTPUT" | grep -q "Usage:"; then
    echo "FAILURE: Usage printed for piped format but shouldn't have"
else
    echo "SUCCESS: Usage NOT printed for piped format"
fi

# Clean up
rm output_format.log output_list.log output_norm.log output_val.log 2>/dev/null
