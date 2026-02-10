package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	rcs "github.com/arran4/golang-rcs"
	"github.com/google/go-cmp/cmp"
)

func TestCircularJson(t *testing.T) {
	input := loadTestInput(t)
	parsedOriginal, err := rcs.ParseFile(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("parse original: %v", err)
	}

	// 1. ToJson logic
	jsonData, err := json.Marshal(parsedOriginal)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	// 2. FromJson logic
	var parsedNew rcs.File
	if err := json.Unmarshal(jsonData, &parsedNew); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	// 3. Compare structs
	if diff := cmp.Diff(parsedOriginal, &parsedNew); diff != "" {
		t.Errorf("Struct mismatch (-want +got):\n%s", diff)
	}

	// 4. Compare string representation
	originalString := parsedOriginal.String()
	newString := parsedNew.String()
	if diff := cmp.Diff(originalString, newString); diff != "" {
		t.Errorf("String representation mismatch (-want +got):\n%s", diff)
	}
}

func TestJsonCommandsStdIO(t *testing.T) {
	// Mock stdout/stdin is tricky with the current implementation which uses os.Stdin/Stdout directly.
	// For this test, we can use a temporary file and the file path argument support.

	input := loadTestInput(t)
	tmpFile, err := ioutil.TempFile("", "rcs_input_*.v")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("remove temp: %v", err)
		}
	}()
	if _, err := tmpFile.Write(input); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close temp: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ToJson(tmpFile.Name())

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy pipe: %v", err)
	}
	jsonOutput := buf.Bytes()

	// Now feed this jsonOutput to FromJson via a temp file
	tmpJsonFile, err := ioutil.TempFile("", "rcs_input_*.json")
	if err != nil {
		t.Fatalf("temp json file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpJsonFile.Name()); err != nil {
			t.Errorf("remove temp json: %v", err)
		}
	}()
	if _, err := tmpJsonFile.Write(jsonOutput); err != nil {
		t.Fatalf("write temp json: %v", err)
	}
	if err := tmpJsonFile.Close(); err != nil {
		t.Fatalf("close temp json: %v", err)
	}

	// Capture stdout again
	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	FromJson(tmpJsonFile.Name())

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	os.Stdout = oldStdout

	var buf2 bytes.Buffer
	if _, err := io.Copy(&buf2, r); err != nil {
		t.Fatalf("copy pipe: %v", err)
	}
	finalOutput := buf2.Bytes()

	// Parse original to string for comparison, as strict byte comparison might fail due to whitespace differences if any,
	// but mostly we want to ensure rcs.ParseFile(input).String() matches finalOutput.
	// Although FromJson calls r.String(), so it should be close.
	// Note: rcs.ParseFile might consume slightly differently than raw bytes.

	parsedOriginal, _ := rcs.ParseFile(bytes.NewReader(input))
	expectedOutput := parsedOriginal.String()

	if diff := cmp.Diff(expectedOutput, string(finalOutput)); diff != "" {
		t.Errorf("Round trip output mismatch (-want +got):\n%s", diff)
	}
}
