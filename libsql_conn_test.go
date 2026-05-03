package main

import "testing"

func TestBuildLibSQLConnectionStringEmptyToken(t *testing.T) {
	got, err := buildLibSQLConnectionString("libsql://example.turso.io", "")
	if err != nil {
		t.Fatalf("buildLibSQLConnectionString() error = %v", err)
	}
	want := "libsql://example.turso.io"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestBuildLibSQLConnectionStringAddsEncodedToken(t *testing.T) {
	got, err := buildLibSQLConnectionString("libsql://example.turso.io", "secret token+/=")
	if err != nil {
		t.Fatalf("buildLibSQLConnectionString() error = %v", err)
	}
	want := "libsql://example.turso.io?authToken=secret+token%2B%2F%3D"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestBuildLibSQLConnectionStringPreservesExistingQuery(t *testing.T) {
	got, err := buildLibSQLConnectionString("libsql://example.turso.io?tls=1", "secret")
	if err != nil {
		t.Fatalf("buildLibSQLConnectionString() error = %v", err)
	}
	want := "libsql://example.turso.io?authToken=secret&tls=1"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestBuildLibSQLConnectionStringInvalidURL(t *testing.T) {
	_, err := buildLibSQLConnectionString("://bad-url", "secret")
	if err == nil {
		t.Fatalf("buildLibSQLConnectionString() error = nil, want error")
	}
}

func TestMaskLibSQLURL(t *testing.T) {
	got := maskLibSQLURL("libsql://example.turso.io?authToken=secret&tls=1")
	want := "libsql://example.turso.io?authToken=***&tls=1"
	if got != want {
		t.Fatalf("maskLibSQLURL() = %q, want %q", got, want)
	}
}

func TestMaskLibSQLURLInvalidInput(t *testing.T) {
	got := maskLibSQLURL("://bad-url")
	want := "<invalid-url>"
	if got != want {
		t.Fatalf("maskLibSQLURL() = %q, want %q", got, want)
	}
}
