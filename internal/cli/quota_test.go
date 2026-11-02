package cli

import (
	"testing"
)

func TestLookupModelLimits(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		tier      string
		wantKnown bool
		wantRPD   int
		wantRPM   int
		wantTPM   int
	}{
		{
			name:      "gemini-2.5-flash free",
			model:     "gemini-2.5-flash",
			tier:      "free",
			wantKnown: true,
			wantRPD:   1500,
			wantRPM:   10,
			wantTPM:   250000,
		},
		{
			name:      "gemini-2.5-flash paid tier 1",
			model:     "gemini-2.5-flash",
			tier:      "1",
			wantKnown: true,
			wantRPD:   10000,
			wantRPM:   1000,
			wantTPM:   1000000,
		},
		{
			name:      "gemini-3.8-flash free",
			model:     "gemini-3.8-flash",
			tier:      "free",
			wantKnown: true,
			wantRPD:   20,
			wantRPM:   5,
			wantTPM:   250000,
		},
		{
			name:      "gemini-3.8-flash-preview paid tier 1 prefix matching",
			model:     "gemini-3.8-flash-preview",
			tier:      "1",
			wantKnown: true,
			wantRPD:   10000,
			wantRPM:   1000,
			wantTPM:   1000000,
		},
		{
			name:      "gemini-4-argon paid tier 1",
			model:     "gemini-4-argon",
			tier:      "1",
			wantKnown: true,
			wantRPD:   1000,
			wantRPM:   150,
			wantTPM:   2000000,
		},
		{
			name:      "gemini-4-argon free tier unavailable",
			model:     "gemini-4-argon",
			tier:      "free",
			wantKnown: false,
		},
		{
			name:      "gemini-4 alias paid tier 1",
			model:     "gemini-4",
			tier:      "1",
			wantKnown: true,
			wantRPD:   1000,
			wantRPM:   150,
			wantTPM:   2000000,
		},
		{
			name:      "unknown model returns false",
			model:     "custom-model",
			tier:      "1",
			wantKnown: false,
		},
		{
			name:      "empty tier returns false",
			model:     "gemini-2.5-flash",
			tier:      "",
			wantKnown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lim, ok := lookupModelLimits(tt.model, tt.tier)
			if ok != tt.wantKnown {
				t.Fatalf("lookupModelLimits(%q, %q) ok = %v, want %v", tt.model, tt.tier, ok, tt.wantKnown)
			}
			if tt.wantKnown {
				if lim.RPD != tt.wantRPD {
					t.Errorf("RPD = %d, want %d", lim.RPD, tt.wantRPD)
				}
				if lim.RPM != tt.wantRPM {
					t.Errorf("RPM = %d, want %d", lim.RPM, tt.wantRPM)
				}
				if lim.TPM != tt.wantTPM {
					t.Errorf("TPM = %d, want %d", lim.TPM, tt.wantTPM)
				}
			}

			rpd := modelRPD(tt.model, tt.tier)
			if tt.wantKnown && rpd != tt.wantRPD {
				t.Errorf("modelRPD = %d, want %d", rpd, tt.wantRPD)
			}
			if !tt.wantKnown && rpd != 0 {
				t.Errorf("modelRPD for unknown = %d, want 0", rpd)
			}

			tpm := lookupModelTPM(tt.model, tt.tier)
			if tt.wantKnown && tpm != tt.wantTPM {
				t.Errorf("lookupModelTPM = %d, want %d", tpm, tt.wantTPM)
			}
			if !tt.wantKnown && tpm != 0 {
				t.Errorf("lookupModelTPM for unknown = %d, want 0", tpm)
			}
		})
	}
}
