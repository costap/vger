package gemini

import (
	"context"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/adapters/cncf"
)

var lookupCNCFDecl = &genai.FunctionDeclaration{
	Name:        "lookup_cncf_project",
	Description: "Look up a technology or project in the live CNCF landscape to get its current graduation stage (graduated, incubating, or sandbox). Use this for every technology you identify — do not rely on training data for CNCF stage.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"name": {
				Type:        genai.TypeString,
				Description: "The technology or project name, e.g. 'Cilium', 'OpenTelemetry', 'Argo CD'",
			},
		},
		Required: []string{"name"},
	},
}

var validateURLDecl = &genai.FunctionDeclaration{
	Name:        "validate_url",
	Description: "Check whether a URL is reachable (returns a valid HTTP response). Always call this before including any URL in the learn_more field.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"url": {
				Type:        genai.TypeString,
				Description: "The URL to validate, e.g. 'https://cilium.io'",
			},
		},
		Required: []string{"url"},
	},
}

// toolSet holds the Go implementations backing the Gemini function declarations.
type toolSet struct {
	cncfClient   *cncf.Client
	declarations []*genai.FunctionDeclaration
}

func newToolSet(cncfClient *cncf.Client) *toolSet {
	return &toolSet{
		cncfClient:   cncfClient,
		declarations: []*genai.FunctionDeclaration{lookupCNCFDecl, validateURLDecl},
	}
}

// execute dispatches a Gemini FunctionCall to the appropriate Go implementation
// and returns the result as a map suitable for a FunctionResponse.
func (t *toolSet) execute(ctx context.Context, fc *genai.FunctionCall) map[string]any {
	switch fc.Name {
	case "lookup_cncf_project":
		name, _ := fc.Args["name"].(string)
		stage, found := t.cncfClient.LookupProject(ctx, name)
		if !found {
			return map[string]any{"found": false, "stage": ""}
		}
		return map[string]any{"found": true, "stage": stage}

	case "validate_url":
		url, _ := fc.Args["url"].(string)
		reachable := t.cncfClient.ValidateURL(ctx, url)
		return map[string]any{"reachable": reachable}

	default:
		return map[string]any{"error": "unknown function: " + fc.Name}
	}
}
