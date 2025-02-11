package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/takutakahashi/kommon/pkg/docker"
)

type Server struct {
	manager *docker.Manager
	// Webhook secret for GitHub
	webhookSecret string
}

type ExecuteRequest struct {
	IssueID string `json:"issue_id"`
	Command string `json:"command"`
}

type CloseRequest struct {
	IssueID string `json:"issue_id"`
}

func NewServer(gooseImage string) (*Server, error) {
	manager, err := docker.NewManager(gooseImage)
	if err != nil {
		return nil, err
	}

	return &Server{
		manager: manager,
	}, nil
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/execute", s.handleExecute)
	http.HandleFunc("/close", s.handleClose)

	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.IssueID == "" || req.Command == "" {
		http.Error(w, "issue_id and command are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Start container if not exists
	if err := s.manager.StartContainer(ctx, req.IssueID); err != nil {
		if err.Error() != fmt.Sprintf("container for issue %s already exists", req.IssueID) {
			http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Execute command
	output, err := s.manager.ExecuteCommand(ctx, req.IssueID, req.Command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute command: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"output": output,
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CloseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.IssueID == "" {
		http.Error(w, "issue_id is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Stop container
	if err := s.manager.StopContainer(ctx, req.IssueID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop container: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Close() error {
	return s.manager.Close()
}
