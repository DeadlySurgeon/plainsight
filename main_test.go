package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/deadlysurgeon/plainsight/auth"
)

func TestRun_Success(t *testing.T) {
	// Fake auth server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":"test-token"}`))
	}))
	defer srv.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := run(ctx, "user", "pass", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "test-token") {
		t.Errorf("expected token printed, got %q", string(out))
	}
}

func TestRun_RequestJWTFailure(t *testing.T) {
	// Fake server that returns bad JSON (will trigger ErrMalformedServerData inside RequestJWT)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":false}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := run(ctx, "user", "pass", srv.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Use errors.Is because RequestJWT wraps
	if !errors.Is(err, auth.ErrMalformedServerData) {
		t.Errorf("expected ErrMalformedServerData, got %v", err)
	}
}

func TestRun_InvalidCreds(t *testing.T) {
	// Missing username
	ctx := context.Background()
	err := run(ctx, "", "pass", "http://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no username provided") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestForTheMoney(t *testing.T) {
	// Fake auth server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":"test-token"}`))
	}))
	defer srv.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Args = append(os.Args, "--username=bob", "--password=password", "--override="+srv.URL)

	main()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "test-token") {
		t.Errorf("expected token printed, got %q", string(out))
	}
}

func TestMain_Exit(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		// Call main() directly, this will os.Exit(1)
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_Exit")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatalf("expected ExitError, got %v", err)
	}
}
