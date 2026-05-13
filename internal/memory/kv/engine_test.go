package kv

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_CreateDir(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "new", "nested")
	eng, err := Open(root)
	if err != nil {
		t.Fatalf("Open should create nested dirs: %v", err)
	}
	defer eng.Close()
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Fatal("Open did not create rootDir")
	}
}

func TestPutGet_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()

	ctx := context.Background()
	key := "wm:bundle:test-001"
	val := []byte(`{"bundle_id":"test-001","goal":"fix provider"}`)

	if err := eng.Put(ctx, key, val); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, err := eng.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(val) {
		t.Fatalf("Get returned unexpected value: got %q, want %q", got, val)
	}
}

func TestGet_NotFound(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()

	_, err = eng.Get(context.Background(), "wm:bundle:missing")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_Tombstone(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()

	ctx := context.Background()
	key := "wm:bundle:to-delete"
	val := []byte(`{"bundle_id":"to-delete"}`)

	if err := eng.Put(ctx, key, val); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := eng.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = eng.Get(ctx, key)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDelete_Reput(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()

	ctx := context.Background()
	key := "wm:bundle:reput"
	val1 := []byte(`{"version":1}`)
	val2 := []byte(`{"version":2}`)

	eng.Put(ctx, key, val1) //nolint:errcheck
	eng.Delete(ctx, key)    //nolint:errcheck
	eng.Put(ctx, key, val2) //nolint:errcheck

	got, err := eng.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after reput: %v", err)
	}
	if string(got) != string(val2) {
		t.Fatalf("expected v2, got %q", got)
	}
}

func TestOpen_ReplayLog(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	// First engine: write, then close.
	eng1, err := Open(dir)
	if err != nil {
		t.Fatalf("Open 1: %v", err)
	}
	eng1.Put(ctx, "wm:bundle:a", []byte(`{"k":"a"}`))
	eng1.Put(ctx, "wm:bundle:b", []byte(`{"k":"b"}`))
	eng1.Delete(ctx, "wm:bundle:a")
	eng1.Close()

	// Second engine: replay on open.
	eng2, err := Open(dir)
	if err != nil {
		t.Fatalf("Open 2: %v", err)
	}
	defer eng2.Close()

	_, err = eng2.Get(ctx, "wm:bundle:a")
	if err != ErrNotFound {
		t.Fatalf("expected a deleted after replay, got %v", err)
	}
	got, err := eng2.Get(ctx, "wm:bundle:b")
	if err != nil {
		t.Fatalf("Get b after replay: %v", err)
	}
	if string(got) != `{"k":"b"}` {
		t.Fatalf("unexpected value for b: %q", got)
	}
}

func TestCompact_Reopen(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	eng1, err := Open(dir)
	if err != nil {
		t.Fatalf("Open 1: %v", err)
	}
	eng1.Put(ctx, "wm:bundle:c1", []byte(`{"v":1}`))
	eng1.Put(ctx, "wm:bundle:c2", []byte(`{"v":2}`))
	if err := eng1.Compact(); err != nil {
		t.Fatalf("Compact: %v", err)
	}
	eng1.Close()

	// Reopen: should load from snapshot + replay zero new log lines.
	eng2, err := Open(dir)
	if err != nil {
		t.Fatalf("Open 2: %v", err)
	}
	defer eng2.Close()

	for _, key := range []string{"wm:bundle:c1", "wm:bundle:c2"} {
		got, err := eng2.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get %q after compact reopen: %v", key, err)
		}
		if len(got) == 0 {
			t.Fatalf("empty value for %q", key)
		}
	}
}

func TestCompact_HistoryPreserved(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	eng.Put(ctx, "wm:bundle:hp", []byte(`{"hp":true}`))
	eng.Compact()
	eng.Close()

	// history.jsonl must still exist and contain the original put.
	logPath := filepath.Join(dir, logFileName)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("history.jsonl should exist after compact: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("history.jsonl was truncated by compact")
	}
}

func TestCorruptSnapshot_FallbackToLogReplay(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	eng1, err := Open(dir)
	if err != nil {
		t.Fatalf("Open 1: %v", err)
	}
	eng1.Put(ctx, "wm:bundle:corrupt-test", []byte(`{"safe":true}`))
	eng1.Compact()
	eng1.Close()

	// Corrupt snapshot header.
	snapPath := filepath.Join(dir, snapshotFileName)
	if err := os.WriteFile(snapPath, []byte("BAD!"), 0640); err != nil {
		t.Fatalf("write corrupt snapshot: %v", err)
	}

	eng2, err := Open(dir)
	if err != nil {
		t.Fatalf("Open 2 after corruption: %v", err)
	}
	defer eng2.Close()

	got, err := eng2.Get(ctx, "wm:bundle:corrupt-test")
	if err != nil {
		t.Fatalf("Get after corrupt snapshot fallback: %v", err)
	}
	if string(got) != `{"safe":true}` {
		t.Fatalf("unexpected value: %q", got)
	}
}

func TestBoundary_KeyEmpty(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()

	if err := eng.Put(context.Background(), "", []byte("x")); err != ErrKeyEmpty {
		t.Fatalf("expected ErrKeyEmpty, got %v", err)
	}
}

func TestBoundary_KeyTooLong(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()

	key := make([]byte, maxKeyLen+1)
	for i := range key {
		key[i] = 'k'
	}
	if err := eng.Put(context.Background(), string(key), []byte("x")); err != ErrKeyTooLong {
		t.Fatalf("expected ErrKeyTooLong, got %v", err)
	}
}

func TestBoundary_ValueTooLong(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()

	val := make([]byte, maxValueLen+1)
	if err := eng.Put(context.Background(), "k", val); err != ErrValueTooLong {
		t.Fatalf("expected ErrValueTooLong, got %v", err)
	}
}

func TestScanPrefix(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()
	ctx := context.Background()

	eng.Put(ctx, "wm:bundle:alpha", []byte(`1`))
	eng.Put(ctx, "wm:bundle:beta", []byte(`2`))
	eng.Put(ctx, "other:key", []byte(`3`))

	it, err := eng.ScanPrefix(ctx, "wm:bundle:")
	if err != nil {
		t.Fatalf("ScanPrefix: %v", err)
	}
	defer it.Close()

	count := 0
	for it.Next() {
		count++
		if it.Key() == "" {
			t.Fatal("empty key in iterator")
		}
		if len(it.Value()) == 0 {
			t.Fatal("empty value in iterator")
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 entries, got %d", count)
	}
}
