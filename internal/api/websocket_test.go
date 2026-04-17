//go:build integration

package api

import (
	"testing"
)

func TestGenerateClientID(t *testing.T) {
	id := generateClientID()
	if id == "" {
		t.Error("generateClientID should return non-empty string")
	}
	if len(id) < 10 {
		t.Errorf("generateClientID returned short ID: %s", id)
	}
	t.Logf("Generated client ID: %s", id)
}

func TestRandomString(t *testing.T) {
	s := randomString(16)
	if len(s) != 16 {
		t.Errorf("randomString(16) returned string of length %d", len(s))
	}

	s2 := randomString(8)
	if len(s2) != 8 {
		t.Errorf("randomString(8) returned string of length %d", len(s2))
	}
}

func TestWebSocketServer_RunLoop(t *testing.T) {
	t.Skip("Run() blocks indefinitely — needs proper shutdown")
}

func TestWebSocketServer_Broadcast_NoClients(t *testing.T) {
	ws := NewWebSocketServer(nil)
	ws.Broadcast(&WebSocketMessage{Type: "test", Payload: map[string]string{"key": "value"}})
	t.Log("Broadcast with no clients completed")
}

func TestWebSocketServer_ShouldSend(t *testing.T) {
	ws := NewWebSocketServer(nil)
	client := &WebSocketClient{id: "test", filters: map[string]bool{"pipeline-1": true}}
	msg := &WebSocketMessage{Type: "update", Payload: "test"}
	result := ws.shouldSend(client, msg)
	t.Logf("shouldSend result: %v", result)
}