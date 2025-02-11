package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takutakahashi/kommon/pkg/agent"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	testClusterName = "kommon-test"
	testNamespace   = "kommon-test"
)

type testEnv struct {
	kubeClient *kubernetes.Clientset
}

func setupTestEnvironment(t *testing.T) (*testEnv, func()) {
	t.Helper()

	// Create kind cluster
	cmd := exec.Command("kind", "create", "cluster", "--name", testClusterName)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to create kind cluster: %s", string(output))

	// Wait for cluster to be ready
	time.Sleep(10 * time.Second)

	// Get kubeconfig
	cmd = exec.Command("kind", "get", "kubeconfig", "--name", testClusterName)
	kubeconfigBytes, err := cmd.Output()
	require.NoError(t, err)

	// Create kubernetes client
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	require.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)

	// Create test namespace
	_, err = clientset.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		},
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	// Load the test image into kind
	cmd = exec.Command("kind", "load", "docker-image", "kommon-agent:latest", "--name", testClusterName)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "failed to load docker image: %s", string(output))

	env := &testEnv{
		kubeClient: clientset,
	}

	// Return cleanup function
	cleanup := func() {
		cmd := exec.Command("kind", "delete", "cluster", "--name", testClusterName)
		if err := cmd.Run(); err != nil {
			t.Logf("Failed to delete test cluster: %v", err)
		}
	}

	return env, cleanup
}

func TestKubernetesExecutor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	env, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create executor options
	opts := ExecutorOptions{
		Type:      ExecutorTypeKubernetes,
		Namespace: testNamespace,
	}

	// Create executor
	executor, err := NewKubernetesExecutor(opts)
	require.NoError(t, err)
	require.NotNil(t, executor)

	// Test initialization
	ctx := context.Background()
	err = executor.Initialize(ctx)
	require.NoError(t, err)

	// Test agent creation
	agentOpts := agent.AgentOptions{
		SessionID: "test-agent-1",
		BaseURL:   "http://localhost:8080",
		APIKey:    "test-key",
	}

	t.Run("CreateAgent", func(t *testing.T) {
		agent, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Verify pod creation
		pod, err := env.kubeClient.CoreV1().Pods(testNamespace).Get(
			ctx,
			fmt.Sprintf("kommon-agent-%s", agentOpts.SessionID),
			metav1.GetOptions{},
		)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("kommon-agent-%s", agentOpts.SessionID), pod.Name)
		assert.Equal(t, testNamespace, pod.Namespace)

		// Try creating duplicate agent
		_, err = executor.CreateAgent(ctx, agentOpts)
		assert.Error(t, err)
	})

	t.Run("ListAgents", func(t *testing.T) {
		agents, err := executor.ListAgents(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 1)
		assert.Contains(t, agents, "test-agent-1")
	})

	t.Run("GetStatus", func(t *testing.T) {
		status, err := executor.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, ExecutorTypeKubernetes, status.Type)
		assert.True(t, status.IsReady)
		assert.Equal(t, 1, status.ActiveAgents)
		assert.NotNil(t, status.ResourceStatus)
	})

	t.Run("DestroyAgent", func(t *testing.T) {
		err := executor.DestroyAgent(ctx, "test-agent-1")
		require.NoError(t, err)

		// Verify pod deletion
		_, err = env.kubeClient.CoreV1().Pods(testNamespace).Get(
			ctx,
			fmt.Sprintf("kommon-agent-%s", agentOpts.SessionID),
			metav1.GetOptions{},
		)
		assert.Error(t, err)

		// Verify agent was removed from executor
		agents, err := executor.ListAgents(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 0)

		// Try destroying non-existent agent
		err = executor.DestroyAgent(ctx, "non-existent")
		assert.Error(t, err)
	})
}
