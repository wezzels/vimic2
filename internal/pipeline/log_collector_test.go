//go:build integration

package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ==================== Log Collector Tests ====================

func setupLogTest(t *testing.T) (*LogCollector, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-log-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("NewPipelineDB failed: %v", err)
	}

	config := &LogConfig{
		StoragePath: filepath.Join(tmpDir, "logs"),
		BufferSize:  100,
	}

	lc, err := NewLogCollector(db, config)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("NewLogCollector failed: %v", err)
	}

	// Set state file to temp directory and clear any state loaded from default path
	lc.SetStateFile(filepath.Join(tmpDir, "logs-state.json"))
	lc.mu.Lock()
	for k := range lc.logs {
		delete(lc.logs, k)
	}
	lc.mu.Unlock()

	t.Cleanup(func() {
		db.Close()
		os.RemoveAll(tmpDir)
	})

	return lc, tmpDir
}

func TestLogCollector_New(t *testing.T) {
	lc, _ := setupLogTest(t)
	if lc == nil {
		t.Fatal("LogCollector should not be nil")
	}
}

func TestLogCollector_CreateLogStream(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	if stream.ID == "" {
		t.Error("Stream ID should not be empty")
	}
	if stream.PipelineID != "test-pipeline" {
		t.Errorf("PipelineID = %s, want test-pipeline", stream.PipelineID)
	}
	if stream.RunnerID != "runner-1" {
		t.Errorf("RunnerID = %s, want runner-1", stream.RunnerID)
	}
	if stream.Stage != "build" {
		t.Errorf("Stage = %s, want build", stream.Stage)
	}
	if !stream.Live {
		t.Error("Stream should be live after creation")
	}
}

func TestLogCollector_WriteLog(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	err = lc.WriteLog(ctx, stream.ID, "info", "Building project...", 0)
	if err != nil {
		t.Fatalf("WriteLog failed: %v", err)
	}

	err = lc.WriteLog(ctx, stream.ID, "error", "Build failed", 1500)
	if err != nil {
		t.Fatalf("WriteLog failed: %v", err)
	}
}

func TestLogCollector_GetLogStream(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	retrieved, err := lc.GetLogStream(stream.ID)
	if err != nil {
		t.Fatalf("GetLogStream failed: %v", err)
	}
	if retrieved.ID != stream.ID {
		t.Errorf("GetLogStream ID = %s, want %s", retrieved.ID, stream.ID)
	}
}

func TestLogCollector_ListLogStreams(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	lc.CreateLogStream(ctx, "test-pipeline", "runner-2", "test", "job-2")

	streams := lc.ListLogStreams("test-pipeline")
	if len(streams) != 2 {
		t.Errorf("ListLogStreams returned %d streams, want 2", len(streams))
	}
}

func TestLogCollector_CloseLogStream(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	err = lc.CloseLogStream(ctx, stream.ID)
	if err != nil {
		t.Fatalf("CloseLogStream failed: %v", err)
	}

	retrieved, err := lc.GetLogStream(stream.ID)
	if err != nil {
		t.Fatalf("GetLogStream failed: %v", err)
	}
	if retrieved.Live {
		t.Error("Stream should not be live after close")
	}
	if retrieved.EndTime == nil {
		t.Error("Stream should have EndTime after close")
	}
}

func TestLogCollector_DeleteLogStream(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	err = lc.DeleteLogStream(ctx, stream.ID)
	if err != nil {
		t.Fatalf("DeleteLogStream failed: %v", err)
	}

	_, err = lc.GetLogStream(stream.ID)
	if err == nil {
		t.Error("GetLogStream should return error for deleted stream")
	}
}

func TestLogCollector_SearchLogs(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	lc.WriteLog(ctx, stream.ID, "info", "Building project successfully", 0)
	lc.WriteLog(ctx, stream.ID, "error", "Build failed with error", 1500)
	lc.WriteLog(ctx, stream.ID, "info", "Running tests", 500)

	// Search for "build"
	results, err := lc.SearchLogs(ctx, "test-pipeline", "build", 10)
	if err != nil {
		t.Fatalf("SearchLogs failed: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("SearchLogs returned %d results, want at least 1", len(results))
	}
}

func TestLogCollector_GetStats(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, _ := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	lc.WriteLog(ctx, stream.ID, "info", "test message", 0)

	stats := lc.GetStats()
	if stats["total_streams"] < 1 {
		t.Errorf("total_streams = %v, want at least 1", stats["total_streams"])
	}
	if stats["live_streams"] < 1 {
		t.Errorf("live_streams = %v, want at least 1", stats["live_streams"])
	}
}

func TestLogCollector_StreamLog(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	// Subscribe to log updates
	ch, err := lc.StreamLog(stream.ID)
	if err != nil {
		t.Fatalf("StreamLog failed: %v", err)
	}

	// Write a log entry
	lc.WriteLog(ctx, stream.ID, "info", "test streaming message", 0)

	// Read from channel (with timeout)
	select {
	case entry := <-ch:
		if entry.Message != "test streaming message" {
			t.Errorf("StreamLog message = %s, want 'test streaming message'", entry.Message)
		}
	case <-time.After(time.Second):
		t.Error("StreamLog timed out waiting for message")
	}
}

func TestLogCollector_DeleteLogsForPipeline(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	lc.CreateLogStream(ctx, "test-pipeline", "runner-2", "test", "job-2")

	err := lc.DeleteLogsForPipeline(ctx, "test-pipeline")
	if err != nil {
		t.Fatalf("DeleteLogsForPipeline failed: %v", err)
	}

	streams := lc.ListLogStreams("test-pipeline")
	if len(streams) != 0 {
		t.Errorf("ListLogStreams returned %d streams after delete, want 0", len(streams))
	}
}

func TestLogCollector_StateFile(t *testing.T) {
	lc, tmpDir := setupLogTest(t)

	stateFile := lc.getStateFile()
	if stateFile == "" {
		t.Error("getStateFile should not return empty string")
	}

	customPath := filepath.Join(tmpDir, "custom-log-state.json")
	lc.SetStateFile(customPath)
	if lc.getStateFile() != customPath {
		t.Errorf("getStateFile = %s, want %s", lc.getStateFile(), customPath)
	}
}

func TestLogCollector_ReadLog(t *testing.T) {
	lc, _ := setupLogTest(t)
	ctx := context.Background()

	stream, err := lc.CreateLogStream(ctx, "test-pipeline", "runner-1", "build", "job-1")
	if err != nil {
		t.Fatalf("CreateLogStream failed: %v", err)
	}

	// Write some logs
	lc.WriteLog(ctx, stream.ID, "info", "Line 1: Starting build", 0)
	lc.WriteLog(ctx, stream.ID, "info", "Line 2: Compiling", 500)
	lc.WriteLog(ctx, stream.ID, "error", "Line 3: Build error", 1000)

	// Read logs
	entries, err := lc.ReadLog(stream.ID, 0, 10)
	if err != nil {
		t.Logf("ReadLog returned error: %v", err)
	}

	t.Logf("ReadLog returned %d entries", len(entries))
}