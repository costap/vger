package gemini

import (
	"context"
	"fmt"

	"github.com/costap/vger/internal/domain"
)

// Client is a stub implementation of domain.VideoAnalyser using the Gemini 2.5 Pro API.
// Replace the stub body with a real google.golang.org/genai call, passing the YouTube URL directly.
type Client struct {
	APIKey string
}

func New(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

func (c *Client) AnalyseVideo(_ context.Context, url string, meta *domain.VideoMetadata) (*domain.Report, error) {
	if url == "" {
		return nil, fmt.Errorf("url must not be empty")
	}
	return &domain.Report{
		VideoTitle: meta.Title,
		VideoURL:   url,
		Summary: "[stub] This talk introduced a novel approach to multi-cluster networking " +
			"using eBPF-based CNI plugins, demonstrating zero-downtime failover across AWS and GCP regions.",
		Technologies: []domain.Technology{
			{
				Name:        "Cilium",
				Description: "eBPF-based Kubernetes CNI and network security platform.",
				WhyRelevant: "Presented as the primary data-plane for cross-cluster traffic management.",
				LearnMore:   "https://cilium.io",
				CNCFStage:   "graduated",
			},
			{
				Name:        "Gateway API",
				Description: "Next-generation Kubernetes traffic routing API.",
				WhyRelevant: "Used to express cross-cluster routing policies declaratively.",
				LearnMore:   "https://gateway-api.sigs.k8s.io",
				CNCFStage:   "",
			},
		},
	}, nil
}
