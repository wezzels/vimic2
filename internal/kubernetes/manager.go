// Package kubernetes provides K8s runner lifecycle management
package kubernetes

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Manager manages Kubernetes-based ephemeral runners
type Manager struct {
	client    *kubernetes.Clientset
	namespace string
	pods      map[string]*PodInfo
}

// PodInfo represents a managed K8s pod runner
type PodInfo struct {
	ID           string       `json:"id"`
	Name        string       `json:"name"`
	Namespace   string       `json:"namespace"`
	Image       string       `json:"image"`
	Platform    string       `json:"platform"`
	Node        string       `json:"node"`
	Status      PodStatus     `json:"status"`
	IPAddress   string       `json:"ip_address"`
	PipelineID  string       `json:"pipeline_id"`
	JobID       string       `json:"job_id"`
	CreatedAt   time.Time    `json:"created_at"`
	Labels      map[string]string `json:"labels"`
}

// PodStatus represents K8s pod status
type PodStatus string

const (
	PodPending   PodStatus = "Pending"
	PodRunning  PodStatus = "Running"
	PodSucceeded PodStatus = "Succeeded"
	PodFailed   PodStatus = "Failed"
)

// Config represents K8s manager configuration
type Config struct {
	KubeConfig    string // path to kubeconfig (empty = in-cluster)
	Namespace     string // namespace for runners
	ImagePullSecrets []string
}

// NewManager creates a new K8s runner manager
func NewManager(config *Config) (*Manager, error) {
	var (
		client *kubernetes.Clientset
		err    error
	)

	if config == nil {
		config = &Config{Namespace: "vimic2-runners"}
	}

	if config.KubeConfig != "" {
		// Use external kubeconfig
		cfg, err := loadKubeConfig(config.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
		client, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create K8s client: %w", err)
		}
	} else {
		// Use in-cluster config
		cfg, err := rest.InClusterConfig()
		if err != nil {
			// Fallback to kind cluster
			cfg = &rest.Config{
				Host: "https://127.0.0.1:6443",
			}
		}
		client, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create K8s client: %w", err)
		}
	}

	return &Manager{
		client:    client,
		namespace: config.Namespace,
		pods:      make(map[string]*PodInfo),
	}, nil
}

// CreatePod creates a new ephemeral K8s pod runner
func (m *Manager) CreatePod(ctx context.Context, opts *PodOptions) (*PodInfo, error) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "vimic2-runner-",
			Namespace: m.namespace,
			Labels: map[string]string{
				"app":              "vimic2",
				"component":        "runner",
				"vimic2.io/pipeline": opts.PipelineID,
				"vimic2.io/job":     opts.JobID,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{{
				Name:  "runner",
				Image: opts.Image,
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    *opts.CPULimit,
						v1.ResourceMemory: *opts.MemoryLimit,
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    *opts.CPURequest,
						v1.ResourceMemory: *opts.MemoryRequest,
					},
				},
				Env: opts.Env,
			}},
			NodeSelector: opts.NodeSelector,
		},
	}

	created, err := m.client.CoreV1().Pods(m.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	info := &PodInfo{
		ID:          created.Name,
		Name:        created.Name,
		Namespace:   m.namespace,
		Image:      opts.Image,
		Platform:   opts.Platform,
		PipelineID: opts.PipelineID,
		JobID:      opts.JobID,
		Status:     PodPending,
		CreatedAt:  time.Now(),
		Labels:     created.Labels,
	}

	m.pods[created.Name] = info
	return info, nil
}

// GetPod returns pod info
func (m *Manager) GetPod(ctx context.Context, name string) (*PodInfo, error) {
	pod, err := m.client.CoreV1().Pods(m.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	info := &PodInfo{
		ID:        pod.Name,
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Image:    pod.Spec.Containers[0].Image,
		Status:   PodStatus(pod.Status.phase),
		IPAddress: pod.Status.PodIP,
		Node:     pod.Spec.NodeName,
		Labels:   pod.Labels,
	}

	if pod.Status.PodIP != "" {
		info.Status = PodRunning
	}

	return info, nil
}

// DeletePod removes a K8s pod runner
func (m *Manager) DeletePod(ctx context.Context, name string) error {
	err := m.client.CoreV1().Pods(m.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	delete(m.pods, name)
	return nil
}

// ListPods returns all runner pods
func (m *Manager) ListPods(ctx context.Context) ([]*PodInfo, error) {
	pods, err := m.client.CoreV1().Pods(m.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=vimic2,component=runner",
	})
	if err != nil {
		return nil, err
	}

	infos := make([]*PodInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		info := &PodInfo{
			ID:        pod.Name,
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Image:    pod.Spec.Containers[0].Image,
			Status:   PodStatus(pod.Status.phase),
			IPAddress: pod.Status.PodIP,
			Node:     pod.Spec.NodeName,
			Labels:   pod.Labels,
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// PodOptions represents pod creation options
type PodOptions struct {
	Image        string
	Platform    string
	PipelineID  string
	JobID       string
	CPULimit    *v1.ResourceList
	MemoryLimit *v1.ResourceList
	CPURequest  *v1.ResourceList
	MemoryRequest *v1.ResourceList
	NodeSelector map[string]string
	Env         []v1.EnvVar
}

func loadKubeConfig(path string) (*rest.Config, error) {
	return rest.ConfigFromKubeConfig([]byte{}), nil
}
