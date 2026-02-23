package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/mur-run/mur-core/internal/session"
)

// ClientMessage is the message format sent from browser to server.
type ClientMessage struct {
	Type      string `json:"type"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Index     int    `json:"index"`
	Field     string `json:"field"`
	Value     any    `json:"value"`
	After     int    `json:"after"`
	Name      string `json:"name"`
	VarType   string `json:"var_type"`
	StepIndex int    `json:"step_index"`
}

// ServerMessage is the message format sent from server to browser.
type ServerMessage struct {
	Type     string                  `json:"type"`
	Workflow *session.AnalysisResult `json:"workflow,omitempty"`
	Message  string                  `json:"message,omitempty"`
	Path     string                  `json:"path,omitempty"`
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	s.addClient(conn)
	defer s.removeClient(conn)

	s.sendState(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var cm ClientMessage
		if err := json.Unmarshal(msg, &cm); err != nil {
			log.Printf("ws parse error: %v", err)
			continue
		}

		s.processMessage(conn, &cm)
	}
}

func (s *Server) addClient(conn *websocket.Conn) {
	s.clientMu.Lock()
	s.clients[conn] = true
	s.clientMu.Unlock()
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.clientMu.Lock()
	delete(s.clients, conn)
	s.clientMu.Unlock()
}

func (s *Server) sendState(conn *websocket.Conn) {
	s.mu.RLock()
	msg := ServerMessage{Type: "state.full", Workflow: s.workflow}
	data, _ := json.Marshal(msg)
	s.mu.RUnlock()

	s.clientMu.Lock()
	conn.WriteMessage(websocket.TextMessage, data)
	s.clientMu.Unlock()
}

func (s *Server) broadcastState() {
	// Caller must hold s.mu
	msg := ServerMessage{Type: "state.full", Workflow: s.workflow}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

func (s *Server) broadcastMessage(msg ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

func (s *Server) sendTo(conn *websocket.Conn, msg ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.clientMu.Lock()
	conn.WriteMessage(websocket.TextMessage, data)
	s.clientMu.Unlock()
}

func (s *Server) processMessage(conn *websocket.Conn, msg *ClientMessage) {
	switch msg.Type {
	case "ai.review":
		go s.handleAIReview(conn, msg)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch msg.Type {
	case "step.reorder":
		s.handleStepReorder(msg)
	case "step.update":
		s.handleStepUpdate(msg)
	case "step.delete":
		s.handleStepDelete(msg)
	case "step.add":
		s.handleStepAdd(msg)
	case "variable.add":
		s.handleVariableAdd(msg)
	case "variable.update":
		s.handleVariableUpdate(msg)
	case "variable.delete":
		s.handleVariableDelete(msg)
	case "workflow.update":
		s.handleWorkflowUpdate(msg)
	default:
		return
	}

	s.broadcastState()
}

func (s *Server) handleStepReorder(msg *ClientMessage) {
	steps := s.workflow.Steps
	from, to := msg.From, msg.To
	if from == to || from < 0 || to < 0 || from >= len(steps) || to >= len(steps) {
		return
	}

	result := make([]session.Step, len(steps))
	copy(result, steps)

	step := result[from]
	if from < to {
		copy(result[from:], result[from+1:to+1])
	} else {
		copy(result[to+1:], result[to:from])
	}
	result[to] = step

	for i := range result {
		result[i].Order = i + 1
	}
	s.workflow.Steps = result
}

func (s *Server) handleStepUpdate(msg *ClientMessage) {
	if msg.Index < 0 || msg.Index >= len(s.workflow.Steps) {
		return
	}
	step := &s.workflow.Steps[msg.Index]

	switch msg.Field {
	case "description":
		if v, ok := msg.Value.(string); ok {
			step.Description = v
		}
	case "command":
		if v, ok := msg.Value.(string); ok {
			step.Command = v
		}
	case "tool":
		if v, ok := msg.Value.(string); ok {
			step.Tool = v
		}
	case "needs_approval":
		if v, ok := msg.Value.(bool); ok {
			step.NeedsApproval = v
		}
	case "on_failure":
		if v, ok := msg.Value.(string); ok {
			step.OnFailure = v
		}
	}
}

func (s *Server) handleStepDelete(msg *ClientMessage) {
	if msg.Index < 0 || msg.Index >= len(s.workflow.Steps) {
		return
	}
	s.workflow.Steps = append(s.workflow.Steps[:msg.Index], s.workflow.Steps[msg.Index+1:]...)
	for i := range s.workflow.Steps {
		s.workflow.Steps[i].Order = i + 1
	}
}

func (s *Server) handleStepAdd(msg *ClientMessage) {
	newStep := session.Step{
		OnFailure: "abort",
	}

	insertAt := msg.After + 1
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(s.workflow.Steps) {
		insertAt = len(s.workflow.Steps)
	}

	steps := make([]session.Step, 0, len(s.workflow.Steps)+1)
	steps = append(steps, s.workflow.Steps[:insertAt]...)
	steps = append(steps, newStep)
	steps = append(steps, s.workflow.Steps[insertAt:]...)

	for i := range steps {
		steps[i].Order = i + 1
	}
	s.workflow.Steps = steps
}

func (s *Server) handleVariableAdd(msg *ClientMessage) {
	v := session.Variable{
		Name:     msg.Name,
		Type:     msg.VarType,
		Required: true,
	}
	if v.Name == "" {
		v.Name = "new_variable"
	}
	if v.Type == "" {
		v.Type = "string"
	}
	s.workflow.Variables = append(s.workflow.Variables, v)
}

func (s *Server) handleVariableUpdate(msg *ClientMessage) {
	if msg.Index < 0 || msg.Index >= len(s.workflow.Variables) {
		return
	}
	v := &s.workflow.Variables[msg.Index]

	switch msg.Field {
	case "name":
		if val, ok := msg.Value.(string); ok {
			v.Name = val
		}
	case "type":
		if val, ok := msg.Value.(string); ok {
			v.Type = val
		}
	case "required":
		if val, ok := msg.Value.(bool); ok {
			v.Required = val
		}
	case "default":
		if val, ok := msg.Value.(string); ok {
			v.Default = val
		}
	case "description":
		if val, ok := msg.Value.(string); ok {
			v.Description = val
		}
	}
}

func (s *Server) handleVariableDelete(msg *ClientMessage) {
	if msg.Index < 0 || msg.Index >= len(s.workflow.Variables) {
		return
	}
	s.workflow.Variables = append(s.workflow.Variables[:msg.Index], s.workflow.Variables[msg.Index+1:]...)
}

func (s *Server) handleWorkflowUpdate(msg *ClientMessage) {
	switch msg.Field {
	case "name":
		if v, ok := msg.Value.(string); ok {
			s.workflow.Name = v
		}
	case "trigger":
		if v, ok := msg.Value.(string); ok {
			s.workflow.Trigger = v
		}
	case "description":
		if v, ok := msg.Value.(string); ok {
			s.workflow.Description = v
		}
	case "tags":
		switch v := msg.Value.(type) {
		case string:
			s.workflow.Tags = splitTags(v)
		case []any:
			tags := make([]string, 0, len(v))
			for _, t := range v {
				if str, ok := t.(string); ok {
					tags = append(tags, str)
				}
			}
			s.workflow.Tags = tags
		}
	case "tools":
		if v, ok := msg.Value.([]any); ok {
			tools := make([]string, 0, len(v))
			for _, t := range v {
				if str, ok := t.(string); ok {
					tools = append(tools, str)
				}
			}
			s.workflow.Tools = tools
		}
	}
}

func (s *Server) handleAIReview(conn *websocket.Conn, msg *ClientMessage) {
	if s.llm == nil {
		s.sendTo(conn, ServerMessage{
			Type:    "ai.suggestion",
			Message: "AI review unavailable — set MUR_API_KEY to enable.",
		})
		return
	}

	s.mu.RLock()
	if msg.StepIndex < 0 || msg.StepIndex >= len(s.workflow.Steps) {
		s.mu.RUnlock()
		return
	}
	step := s.workflow.Steps[msg.StepIndex]
	wfName := s.workflow.Name
	wfDesc := s.workflow.Description
	s.mu.RUnlock()

	prompt := fmt.Sprintf(`Review this step in a workflow and suggest improvements.

Workflow: %s
Description: %s

Step %d: %s
Command: %s
Tool: %s
Needs Approval: %v
On Failure: %s

Suggest a brief improvement (1-3 sentences). Focus on clarity, correctness, and safety.
Respond with just the suggestion text, no JSON or markdown.`,
		wfName, wfDesc, step.Order, step.Description,
		step.Command, step.Tool, step.NeedsApproval, step.OnFailure)

	suggestion, err := s.llm.Complete(prompt)
	if err != nil {
		s.sendTo(conn, ServerMessage{
			Type:    "ai.suggestion",
			Message: fmt.Sprintf("AI review failed: %v", err),
		})
		return
	}

	s.sendTo(conn, ServerMessage{
		Type:    "ai.suggestion",
		Message: strings.TrimSpace(suggestion),
	})
}

func splitTags(s string) []string {
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}
