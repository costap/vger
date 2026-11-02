package gemini

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFlexibleStringSlice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "standard array of strings",
			input:    `["Argo Rollouts", "Flux"]`,
			expected: []string{"Argo Rollouts", "Flux"},
		},
		{
			name: "array of objects with name and description",
			input: `[
				{"name": "Argo Rollouts", "description": "Progressive delivery for Kubernetes"},
				{"name": "Native Kubernetes", "description": "Standard Deployments and StatefulSets"}
			]`,
			expected: []string{
				"Argo Rollouts — Progressive delivery for Kubernetes",
				"Native Kubernetes — Standard Deployments and StatefulSets",
			},
		},
		{
			name: "array of objects with approach and tradeoffs",
			input: `[
				{"approach": "In-place updates", "tradeoffs": "Requires custom CRDs and webhook validation"},
				{"approach": "Pod recreate", "tradeoffs": "Service disruption on rollouts"}
			]`,
			expected: []string{
				"In-place updates — Requires custom CRDs and webhook validation",
				"Pod recreate — Service disruption on rollouts",
			},
		},
		{
			name: "array of objects with speaker and affiliation",
			input: `[
				{"speaker": "Jane Doe", "affiliation": "Red Hat"},
				{"speaker": "John Smith"}
			]`,
			expected: []string{
				"Jane Doe (Red Hat)",
				"John Smith",
			},
		},
		{
			name: "mixed array of strings and objects",
			input: `[
				"Standard Kubernetes Deployments",
				{"name": "Argo Rollouts", "description": "Canary rollouts"}
			]`,
			expected: []string{
				"Standard Kubernetes Deployments",
				"Argo Rollouts — Canary rollouts",
			},
		},
		{
			name:     "single string",
			input:    `"Argo Rollouts"`,
			expected: []string{"Argo Rollouts"},
		},
		{
			name:     "single object",
			input:    `{"name": "Argo Rollouts", "description": "Canary controller"}`,
			expected: []string{"Argo Rollouts — Canary controller"},
		},
		{
			name:     "null",
			input:    `null`,
			expected: nil,
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: []string{},
		},
		{
			name: "array of objects with nested tradeoffs slice",
			input: `[
				{
					"approach": "Argo Rollouts",
					"tradeoffs": ["Requires CRDs", "No in-place update support"]
				}
			]`,
			expected: []string{
				"Argo Rollouts — Requires CRDs, No in-place update support",
			},
		},
		{
			name: "array of objects with single arbitrary key-value",
			input: `[
				{"Argo Rollouts": "Progressive delivery"}
			]`,
			expected: []string{
				"Argo Rollouts — Progressive delivery",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result flexibleStringSlice
			err := json.Unmarshal([]byte(tt.input), &result)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual([]string(result), tt.expected) {
				t.Errorf("got %v, want %v", []string(result), tt.expected)
			}
		})
	}
}

func TestResearchReportJSON_UnmarshalWithCompetingApproachesObjects(t *testing.T) {
	// Reconstruct the exact failure scenario from research "OpenKruise"
	input := `{
		"brief": "OpenKruise is a CNCF incubating suite of operators that extends native Kubernetes with advanced workload management primitives, focusing on non-disruptive operations and high-density lifecycle management.",
		"landscape_map": [
			{
				"name": "OpenKruise",
				"cncf_stage": "incubating",
				"category": "Workload Management",
				"homepage": "https://openkruise.io",
				"relevance": "Core topic"
			}
		],
		"evidence_videos": [],
		"investigation_paths": [
			{
				"title": "Start with fundamentals",
				"description": "Learn clone sets and in-place updates",
				"actions": [
					{"step": "Read documentation"},
					{"step": "Install via Helm"}
				]
			}
		],
		"competing_approaches": [
			{
				"name": "Native Kubernetes Workloads",
				"description": "Standard Deployments and StatefulSets without CRDs"
			},
			{
				"name": "Argo Rollouts",
				"description": "Progressive delivery for Kubernetes"
			}
		],
		"verdict": "Invest time in OpenKruise if you need in-place updates and advanced rollouts."
	}`

	var r researchReportJSON
	if err := json.Unmarshal([]byte(input), &r); err != nil {
		t.Fatalf("failed to unmarshal research report: %v", err)
	}

	if len(r.CompetingApproaches) != 2 {
		t.Fatalf("expected 2 competing approaches, got %d", len(r.CompetingApproaches))
	}
	expectedApproach0 := "Native Kubernetes Workloads — Standard Deployments and StatefulSets without CRDs"
	if r.CompetingApproaches[0] != expectedApproach0 {
		t.Errorf("got %q, want %q", r.CompetingApproaches[0], expectedApproach0)
	}

	if len(r.InvestigationPaths) != 1 {
		t.Fatalf("expected 1 investigation path, got %d", len(r.InvestigationPaths))
	}
	if len(r.InvestigationPaths[0].Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(r.InvestigationPaths[0].Actions))
	}
	if r.InvestigationPaths[0].Actions[0] != "Read documentation" {
		t.Errorf("got %q, want 'Read documentation'", r.InvestigationPaths[0].Actions[0])
	}
}
