package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/takutakahashi/kommon/pkg/agent"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesExecutor struct {
	client    *kubernetes.Clientset
	namespace string
	agents    map[string]bool
	mu        sync.RWMutex
}

type KubernetesAgent struct {
	sessionID string
	client    *kubernetes.Clientset
	namespace string
}

func (a *KubernetesAgent) GetSessionID() string {
	return a.sessionID
}

func (a *KubernetesAgent) StartSession(ctx context.Context) error {
	// For now, just verify that the pod exists
	_, err := a.client.CoreV1().Pods(a.namespace).Get(ctx, fmt.Sprintf("kommon-agent-%s", a.sessionID), metav1.GetOptions{})
	return err
}

func (a *KubernetesAgent) Execute(ctx context.Context, input string) (string, error) {
	// TODO: Implement execution logic
	// This could involve sending commands to the pod or reading its logs
	return "", nil
}

func NewKubernetesExecutor(opts ExecutorOptions) (Executor, error) {
	var (
		config *rest.Config
		err    error
	)

	// Try to use in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	namespace := opts.Namespace
	if namespace == "" {
		namespace = "default"
	}

	return &KubernetesExecutor{
		client:    clientset,
		namespace: namespace,
		agents:    make(map[string]bool),
	}, nil
}

func (e *KubernetesExecutor) Initialize(ctx context.Context) error {
	// Check if namespace exists
	_, err := e.client.CoreV1().Namespaces().Get(ctx, e.namespace, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check namespace: %v", err)
	}

	if errors.IsNotFound(err) {
		// Create namespace if it doesn't exist
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: e.namespace,
			},
		}
		_, err = e.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create namespace: %v", err)
		}
	}

	return nil
}

func (e *KubernetesExecutor) CreateAgent(ctx context.Context, opts agent.GooseOptions) (agent.Agent, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.agents[opts.SessionID] {
		return nil, fmt.Errorf("agent with session ID %s already exists", opts.SessionID)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("kommon-agent-%s", opts.SessionID),
			Labels: map[string]string{
				"app":        "kommon",
				"component":  "agent",
				"session-id": opts.SessionID,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "agent",
					Image: "kommon-agent:latest", // TODO: Make configurable
					Env: []corev1.EnvVar{
						{
							Name:  "SESSION_ID",
							Value: opts.SessionID,
						},
						{
							Name:  "API_KEY",
							Value: opts.APIKey,
						},
					},
				},
			},
		},
	}

	_, err := e.client.CoreV1().Pods(e.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent pod: %v", err)
	}

	e.agents[opts.SessionID] = true

	return &KubernetesAgent{
		sessionID: opts.SessionID,
		client:    e.client,
		namespace: e.namespace,
	}, nil
}

func (e *KubernetesExecutor) ListAgents(ctx context.Context) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	agents := make([]string, 0, len(e.agents))
	for id := range e.agents {
		agents = append(agents, id)
	}
	return agents, nil
}

func (e *KubernetesExecutor) DestroyAgent(ctx context.Context, sessionID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.agents[sessionID] {
		return fmt.Errorf("agent with session ID %s does not exist", sessionID)
	}

	podName := fmt.Sprintf("kommon-agent-%s", sessionID)
	err := e.client.CoreV1().Pods(e.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete agent pod: %v", err)
	}

	delete(e.agents, sessionID)
	return nil
}

func (e *KubernetesExecutor) GetStatus(ctx context.Context) (*ExecutorStatus, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check if Kubernetes API is accessible
	_, err := e.client.CoreV1().Namespaces().Get(ctx, e.namespace, metav1.GetOptions{})
	if err != nil {
		return &ExecutorStatus{
			Type:           ExecutorTypeKubernetes,
			IsReady:        false,
			ActiveAgents:   len(e.agents),
			ResourceStatus: nil,
		}, nil
	}

	return &ExecutorStatus{
		Type:           ExecutorTypeKubernetes,
		IsReady:        true,
		ActiveAgents:   len(e.agents),
		ResourceStatus: &ResourceStatus{},
	}, nil
}
