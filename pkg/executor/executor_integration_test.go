package executor

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takutakahashi/kommon/pkg/agent"
)

func TestLocalExecutorIntegration(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "local-executor-integration-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Executor の作成
	opts := ExecutorOptions{
		Type:      ExecutorTypeLocal,
		ConfigDir: tmpDir,
	}

	executor, err := NewLocalExecutor(opts)
	require.NoError(t, err)
	require.NotNil(t, executor)

	ctx := context.Background()
	err = executor.Initialize(ctx)
	require.NoError(t, err)

	// テスト後のクリーンアップを確実に実行
	defer func() {
		// すべてのエージェントを削除
		agents, _ := executor.ListAgents(ctx)
		for _, id := range agents {
			_ = executor.DestroyAgent(ctx, id)
		}
		_ = executor.Close()
	}()

	// 複数のエージェントを使用したテストケース
	t.Run("MultipleAgentsTest", func(t *testing.T) {
		// 複数のエージェントを作成
		agentIDs := []string{"agent-1", "agent-2", "agent-3"}
		for _, id := range agentIDs {
			agentOpts := agent.AgentOptions{
				SessionID: id,
				BaseURL:   "http://localhost:8080",
				APIKey:    "test-key",
			}
			agent, err := executor.CreateAgent(ctx, agentOpts)
			require.NoError(t, err)
			require.NotNil(t, agent)
		}

		// エージェントリストを確認
		agents, err := executor.ListAgents(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, len(agentIDs))
		for _, id := range agentIDs {
			assert.Contains(t, agents, id)
		}

		// ステータスを確認
		status, err := executor.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, len(agentIDs), status.ActiveAgents)
		assert.True(t, status.IsReady)

		// エージェントを1つずつ削除
		for _, id := range agentIDs {
			err := executor.DestroyAgent(ctx, id)
			require.NoError(t, err)

			// 削除後のエージェント数を確認
			remainingAgents, err := executor.ListAgents(ctx)
			require.NoError(t, err)
			assert.NotContains(t, remainingAgents, id)
		}
	})

	// リソース使用状況のテスト
	t.Run("ResourceStatusTest", func(t *testing.T) {
		status, err := executor.GetStatus(ctx)
		require.NoError(t, err)
		
		// リソース情報が取得できることを確認
		assert.NotNil(t, status.ResourceStatus)
		assert.GreaterOrEqual(t, status.ResourceStatus.CPUUsage, float64(0))
		assert.GreaterOrEqual(t, status.ResourceStatus.MemoryUsage, float64(0))
		assert.GreaterOrEqual(t, status.ResourceStatus.DiskUsage, float64(0))
	})

	// エラー処理のテスト
	t.Run("ErrorHandlingTest", func(t *testing.T) {
		// 存在しないエージェントの削除
		err := executor.DestroyAgent(ctx, "non-existent-agent")
		assert.Error(t, err)

		// 重複するエージェントの作成
		agentOpts := agent.AgentOptions{
			SessionID: "duplicate-agent",
			BaseURL:   "http://localhost:8080",
			APIKey:    "test-key",
		}

		agent1, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent1)

		// 同じIDで2回目の作成を試みる
		agent2, err := executor.CreateAgent(ctx, agentOpts)
		assert.Error(t, err)
		assert.Nil(t, agent2)

		// テスト後にエージェントを削除
		err = executor.DestroyAgent(ctx, "duplicate-agent")
		require.NoError(t, err)
	})
}

func TestLocalExecutorCleanup(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "local-executor-cleanup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Executor の作成
	opts := ExecutorOptions{
		Type:      ExecutorTypeLocal,
		ConfigDir: tmpDir,
	}

	executor, err := NewLocalExecutor(opts)
	require.NoError(t, err)
	require.NotNil(t, executor)

	ctx := context.Background()
	err = executor.Initialize(ctx)
	require.NoError(t, err)

	// エージェントを作成
	agentOpts := agent.AgentOptions{
		SessionID: "cleanup-test-agent",
		BaseURL:   "http://localhost:8080",
		APIKey:    "test-key",
	}

	agent, err := executor.CreateAgent(ctx, agentOpts)
	require.NoError(t, err)
	require.NotNil(t, agent)

	// エージェントが作成されていることを確認
	agents, err := executor.ListAgents(ctx)
	require.NoError(t, err)
	assert.Contains(t, agents, "cleanup-test-agent")

	// Executorをクローズ
	err = executor.Close()
	require.NoError(t, err)

	// すべてのエージェントが削除されていることを確認
	agents, err = executor.ListAgents(ctx)
	require.NoError(t, err)
	assert.Empty(t, agents)
}