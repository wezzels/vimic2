// Package pipeline provides log collection
package pipeline

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogCollector collects and stores logs
type LogCollector struct {
	db          *PipelineDB
	storagePath string
	stateFile   string
	logs        map[string]*LogStream
	mu          sync.RWMutex
	bufferSize  int
}

// LogStream represents a log stream
type LogStream struct {
	ID          string             `json:"id"`
	PipelineID  string             `json:"pipeline_id"`
	RunnerID    string             `json:"runner_id"`
	Stage       string             `json:"stage"`
	JobID       string             `json:"job_id"`
	LineCount   int                `json:"line_count"`
	Offset      int64              `json:"offset"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     *time.Time         `json:"end_time,omitempty"`
	Live        bool               `json:"live"`
	Subscribers []chan<- *LogEntry `json:"-"`
	file        *os.File
}

// LogEntry represents a log line
type LogEntry struct {
	ID         string    `json:"id"`
	PipelineID string    `json:"pipeline_id"`
	RunnerID   string    `json:"runner_id"`
	Stage      string    `json:"stage"`
	JobID      string    `json:"job_id"`
	Timestamp  time.Time `json:"timestamp"`
	Level      string    `json:"level"`
	Message    string    `json:"message"`
	Duration   int64     `json:"duration_ms"`
}

// LogConfig represents log collector configuration
type LogConfig struct {
	StoragePath string `json:"storage_path"`
	BufferSize  int    `json:"buffer_size"`
}

// NewLogCollector creates a new log collector
func NewLogCollector(db *PipelineDB, config *LogConfig) (*LogCollector, error) {
	if config == nil {
		config = &LogConfig{
			StoragePath: filepath.Join(os.Getenv("HOME"), ".vimic2", "logs"),
			BufferSize:  1000,
		}
	}

	lc := &LogCollector{
		db:          db,
		storagePath: config.StoragePath,
		logs:        make(map[string]*LogStream),
		bufferSize:  config.BufferSize,
	}

	// Create storage directory
	if err := os.MkdirAll(lc.storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Load state
	if err := lc.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return lc, nil
}

// loadState loads log state from disk
func (lc *LogCollector) loadState() error {
	stateFile := lc.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var logs []*LogStream
	if err := json.Unmarshal(data, &logs); err != nil {
		return err
	}

	for _, log := range logs {
		lc.logs[log.ID] = log
	}

	return nil
}

// saveState saves log state to disk
// NOTE: Caller must hold lc.mu lock before calling this.
func (lc *LogCollector) saveState() error {
	logs := make([]*LogStream, 0, len(lc.logs))
	for _, log := range lc.logs {
		logs = append(logs, log)
	}

	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}

	stateFile := lc.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (lc *LogCollector) getStateFile() string {
	if lc.stateFile != "" {
		return lc.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "logs-state.json")
}

// SetStateFile sets the state file path
func (lc *LogCollector) SetStateFile(path string) {
	lc.stateFile = path
}

// CreateLogStream creates a new log stream
func (lc *LogCollector) CreateLogStream(ctx context.Context, pipelineID, runnerID, stage, jobID string) (*LogStream, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Generate stream ID
	streamID := generateStreamID()

	// Create log directory
	logDir := filepath.Join(lc.storagePath, pipelineID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", streamID))
	file, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Create stream
	stream := &LogStream{
		ID:         streamID,
		PipelineID: pipelineID,
		RunnerID:   runnerID,
		Stage:      stage,
		JobID:      jobID,
		LineCount:  0,
		Offset:     0,
		StartTime:  time.Now(),
		Live:       true,
		file:       file,
	}

	lc.logs[streamID] = stream

	// Save state
	if err := lc.saveState(); err != nil {
		file.Close()
		os.Remove(logPath)
		delete(lc.logs, streamID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return stream, nil
}

// WriteLog writes a log entry
func (lc *LogCollector) WriteLog(ctx context.Context, streamID, level, message string, duration int64) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return fmt.Errorf("log stream not found: %s", streamID)
	}

	// Create entry
	entry := &LogEntry{
		ID:         generateLogID(),
		PipelineID: stream.PipelineID,
		RunnerID:   stream.RunnerID,
		Stage:      stream.Stage,
		JobID:      stream.JobID,
		Timestamp:  time.Now(),
		Level:      level,
		Message:    message,
		Duration:   duration,
	}

	// Format log line
	logLine := fmt.Sprintf("%s [%s] %s", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
	if duration > 0 {
		logLine = fmt.Sprintf("%s (%dms)", logLine, duration)
	}
	logLine = logLine + "\n"

	// Write to file
	if stream.file != nil {
		n, err := stream.file.WriteString(logLine)
		if err != nil {
			return fmt.Errorf("failed to write log: %w", err)
		}
		stream.Offset += int64(n)
		stream.LineCount++
	}

	// Save to database
	if err := lc.db.SaveLog(ctx, entry); err != nil {
		return fmt.Errorf("failed to save log: %w", err)
	}

	// Notify subscribers
	for _, sub := range stream.Subscribers {
		select {
		case sub <- entry:
		default:
			// Channel full, skip
		}
	}

	return nil
}

// ReadLog reads logs from a stream
func (lc *LogCollector) ReadLog(streamID string, offset int64, limit int) ([]*LogEntry, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return nil, fmt.Errorf("log stream not found: %s", streamID)
	}

	if stream.file == nil {
		return nil, fmt.Errorf("log file not open")
	}

	// Seek to offset
	if _, err := stream.file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Read from start to find lines
	if offset > 0 {
		// Skip to offset position
		if _, err := stream.file.Seek(offset, 0); err != nil {
			return nil, fmt.Errorf("failed to seek to offset: %w", err)
		}
	}

	// Read lines using bufio.Scanner
	scanner := io.LimitReader(stream.file, 1<<20) // 1 MB limit
	bufScanner := bufio.NewScanner(scanner)
	
	entries := make([]*LogEntry, 0, limit)
	linesRead := 0
	
	for bufScanner.Scan() && (limit == 0 || linesRead < limit) {
		line := bufScanner.Text()
		if line == "" {
			continue
		}
		
		// Parse log line format: 2026-04-14T12:00:00Z [INFO] message
		entry := &LogEntry{}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 3 {
			if t, err := time.Parse(time.RFC3339, parts[0]); err == nil {
				entry.Timestamp = t
			}
			level := strings.Trim(parts[1], "[]")
			entry.Level = level
			entry.Message = parts[2]
		}
		
		entries = append(entries, entry)
		linesRead++
	}

	return entries, bufScanner.Err()
}

// StreamLog subscribes to live log updates
func (lc *LogCollector) StreamLog(streamID string) (<-chan *LogEntry, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return nil, fmt.Errorf("log stream not found: %s", streamID)
	}

	if !stream.Live {
		return nil, fmt.Errorf("log stream is not live")
	}

	// Create subscriber channel
	ch := make(chan *LogEntry, lc.bufferSize)
	stream.Subscribers = append(stream.Subscribers, ch)

	return ch, nil
}

// Unsubscribe unsubscribes from log updates
func (lc *LogCollector) Unsubscribe(streamID string, ch chan<- *LogEntry) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return fmt.Errorf("log stream not found: %s", streamID)
	}

	// Remove subscriber - cast to compare channel pointers
	for i, sub := range stream.Subscribers {
		// Compare channel pointers by casting to same type
		if interface{}(sub) == interface{}(ch) {
			stream.Subscribers = append(stream.Subscribers[:i], stream.Subscribers[i+1:]...)
			break
		}
	}

	return nil
}

// CloseLogStream closes a log stream
func (lc *LogCollector) CloseLogStream(ctx context.Context, streamID string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return fmt.Errorf("log stream not found: %s", streamID)
	}

	// Close file
	if stream.file != nil {
		stream.file.Close()
		stream.file = nil
	}

	// Mark as ended
	now := time.Time{}
	stream.EndTime = &now
	stream.Live = false

	// Close subscribers
	for _, sub := range stream.Subscribers {
		close(sub)
	}
	stream.Subscribers = nil

	// Save state
	return lc.saveState()
}

// GetLogStream returns a log stream by ID
func (lc *LogCollector) GetLogStream(streamID string) (*LogStream, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return nil, fmt.Errorf("log stream not found: %s", streamID)
	}

	return stream, nil
}

// ListLogStreams lists log streams for a pipeline
func (lc *LogCollector) ListLogStreams(pipelineID string) []*LogStream {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	streams := make([]*LogStream, 0)
	for _, stream := range lc.logs {
		if stream.PipelineID == pipelineID {
			streams = append(streams, stream)
		}
	}

	return streams
}

// DeleteLogStream deletes a log stream
func (lc *LogCollector) DeleteLogStream(ctx context.Context, streamID string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	stream, ok := lc.logs[streamID]
	if !ok {
		return fmt.Errorf("log stream not found: %s", streamID)
	}

	// Close file if open
	if stream.file != nil {
		stream.file.Close()
	}

	// Delete log file
	logPath := filepath.Join(lc.storagePath, stream.PipelineID, fmt.Sprintf("%s.log", streamID))
	os.Remove(logPath)

	// Delete from memory
	delete(lc.logs, streamID)

	// Save state
	return lc.saveState()
}

// DeleteLogsForPipeline deletes all logs for a pipeline
func (lc *LogCollector) DeleteLogsForPipeline(ctx context.Context, pipelineID string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	for id, stream := range lc.logs {
		if stream.PipelineID == pipelineID {
			if stream.file != nil {
				stream.file.Close()
			}
			delete(lc.logs, id)
		}
	}

	// Remove directory
	dir := filepath.Join(lc.storagePath, pipelineID)
	os.RemoveAll(dir)

	// Save state
	return lc.saveState()
}

// SearchLogs searches logs by query
func (lc *LogCollector) SearchLogs(ctx context.Context, pipelineID, query string, limit int) ([]*LogEntry, error) {
	// Get logs from database
	logs, err := lc.db.ListLogsByPipeline(ctx, pipelineID, limit*10, 0) // Get more for filtering
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	// Filter by query - case-insensitive search
	query = strings.ToLower(query)
	entries := make([]*LogEntry, 0, limit)
	
	for _, log := range logs {
		// Check message, level, stage
		if strings.Contains(strings.ToLower(log.Message), query) ||
			strings.Contains(strings.ToLower(log.Level), query) ||
			strings.Contains(strings.ToLower(log.Stage), query) {
			entries = append(entries, log)
			if len(entries) >= limit {
				break
			}
		}
	}

	return entries, nil
}

// GetStats returns log statistics
func (lc *LogCollector) GetStats() map[string]int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var totalLines int
	var totalSize int64
	liveStreams := 0

	for _, stream := range lc.logs {
		totalLines += stream.LineCount
		totalSize += stream.Offset
		if stream.Live {
			liveStreams++
		}
	}

	return map[string]int{
		"total_streams": len(lc.logs),
		"live_streams":  liveStreams,
		"total_lines":   totalLines,
		"total_size":    int(totalSize),
	}
}

// Helper functions

func generateStreamID() string {
	return fmt.Sprintf("stream-%s-%d", randomString(8), time.Now().Unix())
}

func generateLogID() string {
	return fmt.Sprintf("log-%s-%d", randomString(8), time.Now().UnixNano())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
