package agent

import (
	"fmt"
	"sync"

	"github.com/takutakahashi/kommon/pkg/agent"
)

// Manager は Agent インスタンスを管理します
type Manager struct {
	agents map[string]agent.Agent
	mu     sync.RWMutex
}

// NewManager は新しい Manager インスタンスを作成します
func NewManager() *Manager {
	return &Manager{
		agents: make(map[string]agent.Agent),
	}
}

// sessionID は一意のセッションIDを生成します
func sessionID(repoFullName string, issueNumber int) string {
	return fmt.Sprintf("%s-%d", repoFullName, issueNumber)
}

// GetAgent は Agent インスタンスを取得または作成します
func (m *Manager) GetAgent(repoFullName string, issueNumber int, installationToken string) agent.Agent {
	m.mu.Lock()
	defer m.mu.Unlock()

	sID := sessionID(repoFullName, issueNumber)
	if _, ok := m.agents[sID]; !ok {
		m.agents[sID] = &agent.GooseAgent{
			Opts: agent.GooseOptions{
				SessionID:   sID,
				APIType:     agent.GooseAPITypeOpenRouter,
				APIKey:      installationToken,
				Instruction: "You are a helpful assistant that can answer questions and help with tasks.",
			},
		}
	}
	return m.agents[sID]
}

// RemoveAgent は Agent インスタンスを削除します
func (m *Manager) RemoveAgent(repoFullName string, issueNumber int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.agents, sessionID(repoFullName, issueNumber))
}