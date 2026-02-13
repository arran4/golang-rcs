package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
	input := loadTestInput(t)
	tmpFile, err := os.CreateTemp("", "rcs_input_*.v")
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

	// Act: ToJson using "-" as file (stdin simulation is complex here as ToJson takes filename "-")
	// Actually ToJson("-") reads from os.Stdin. So we need to mock os.Stdin too if we use "-".
	// But in this test we can test "file -> stdout" mode if we pass filename but no -o.
	// Wait, ToJson(file) writes to file + .json by default.
	// ToJson("-") writes to stdout.
	// Let's test "File -> File" default behavior first in another test.
	// Here let's test Stdin -> Stdout behavior.

	// Mock Stdin
	oldStdin := os.Stdin
	rIn, wIn, _ := os.Pipe()
	os.Stdin = rIn

	go func() {
		defer func() {
			if err := wIn.Close(); err != nil {
				t.Errorf("close input pipe: %v", err)
			}
		}()
		if _, err := wIn.Write(input); err != nil {
			t.Errorf("write input pipe: %v", err)
		}
	}()

	ToJson("", false, "-")

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy pipe: %v", err)
	}
	jsonOutput := buf.Bytes()

	// Now feed this jsonOutput to FromJson via Stdin -> Stdout
	oldStdout = os.Stdout
	oldStdin = os.Stdin
	r, w, _ = os.Pipe()
	rIn, wIn, _ = os.Pipe()
	os.Stdout = w
	os.Stdin = rIn

	go func() {
		defer func() {
			if err := wIn.Close(); err != nil {
				t.Errorf("close input pipe: %v", err)
			}
		}()
		if _, err := wIn.Write(jsonOutput); err != nil {
			t.Errorf("write input pipe: %v", err)
		}
	}()

	FromJson("", false, "-")

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	var buf2 bytes.Buffer
	if _, err := io.Copy(&buf2, r); err != nil {
		t.Fatalf("copy pipe: %v", err)
	}
	finalOutput := buf2.Bytes()

	parsedOriginal, _ := rcs.ParseFile(bytes.NewReader(input))
	expectedOutput := parsedOriginal.String()

	if diff := cmp.Diff(expectedOutput, string(finalOutput)); diff != "" {
		t.Errorf("Round trip output mismatch (-want +got):\n%s", diff)
	}
}

func TestJsonCommandsFileToFile(t *testing.T) {
	dir := t.TempDir()
	input := loadTestInput(t)
	inputFile := filepath.Join(dir, "input,v")
	if err := os.WriteFile(inputFile, input, 0644); err != nil {
		t.Fatal(err)
	}

	// 1. ToJson default output
	ToJson("", false, inputFile)
	expectedJsonFile := inputFile + ".json"
	if _, err := os.Stat(expectedJsonFile); os.IsNotExist(err) {
		t.Fatalf("Expected output file %s does not exist", expectedJsonFile)
	}

	// 2. FromJson default output
	// Need to handle potential overwrite issue if we write back to input,v?
	// FromJson writes to trimmed suffix. input,v.json -> input,v
	// input,v already exists. Should fail without force.

	// Capture log panic? Testing log.Panicf is hard without capturing output or recovering.
	// Let's assume user uses a library that allows error return instead of panic for CLI,
	// but currently it panics. We can use defer recover.

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic due to existing output file without force")
			}
		}()
		FromJson("", false, expectedJsonFile)
	}()

	// 3. FromJson with force
	FromJson("", true, expectedJsonFile)
	// Verify content matches original
	content, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatal(err)
	}
	parsedOriginal, _ := rcs.ParseFile(bytes.NewReader(input))
	// input might vary slightly due to whitespace, compare via rcs struct or string()
	// parsedOriginal.String() is the normalized output.
	// But `content` is what FromJson wrote.
	if diff := cmp.Diff(parsedOriginal.String(), string(content)); diff != "" {
		t.Errorf("File to File content mismatch:\n%s", diff)
	}

	// 4. Custom output
	customOut := filepath.Join(dir, "custom.json")
	ToJson(customOut, false, inputFile)
	if _, err := os.Stat(customOut); os.IsNotExist(err) {
		t.Fatalf("Expected custom output file %s", customOut)
	}
}
