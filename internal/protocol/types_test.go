package protocol

import "testing"

func TestRequestValidateAcceptsMinimalValidRequest(t *testing.T) {
	req := Request{
		Version:   VersionV1,
		RequestID: "req-123",
		Script:    "homecloud/site/apply",
		Argv:      []string{"sites/homes/ruyi/hass"},
		Env: map[string]string{
			"TARGET": "server",
		},
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}
}

func TestRequestValidateRejectsMissingVersion(t *testing.T) {
	req := Request{RequestID: "req-123", Script: "homecloud/site/apply"}

	if err := req.Validate(); err != ErrMissingVersion {
		t.Fatalf("expected ErrMissingVersion, got %v", err)
	}
}

func TestRequestValidateRejectsUnsupportedVersion(t *testing.T) {
	req := Request{Version: "v2", RequestID: "req-123", Script: "homecloud/site/apply"}

	if err := req.Validate(); err == nil || err.Error() != "unsupported version: v2" {
		t.Fatalf("expected unsupported version error, got %v", err)
	}
}

func TestRequestValidateRejectsMissingRequestID(t *testing.T) {
	req := Request{Version: VersionV1, Script: "homecloud/site/apply"}

	if err := req.Validate(); err != ErrMissingRequestID {
		t.Fatalf("expected ErrMissingRequestID, got %v", err)
	}
}

func TestRequestValidateRejectsReservedEnvKey(t *testing.T) {
	req := Request{
		Version:   VersionV1,
		RequestID: "req-123",
		Script:    "homecloud/site/apply",
		Env: map[string]string{
			"RR_TARGET": "server",
		},
	}

	if err := req.Validate(); err == nil || err.Error() != "reserved environment key conflict: RR_TARGET" {
		t.Fatalf("expected reserved env conflict, got %v", err)
	}
}

func TestValidateScriptPath(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr error
	}{
		{name: "valid", script: "homecloud/site/apply", wantErr: nil},
		{name: "missing", script: "", wantErr: ErrMissingScript},
		{name: "absolute", script: "/bin/sh", wantErr: ErrInvalidScriptPath},
		{name: "parent traversal", script: "homecloud/../secret", wantErr: ErrInvalidScriptPath},
		{name: "dot segment", script: "./apply", wantErr: ErrInvalidScriptPath},
		{name: "double slash", script: "homecloud//apply", wantErr: ErrInvalidScriptPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptPath(tt.script)
			if err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestIsReservedEnvKey(t *testing.T) {
	if !IsReservedEnvKey("RR_TARGET") {
		t.Fatal("expected RR_TARGET to be reserved")
	}
	if IsReservedEnvKey("TARGET") {
		t.Fatal("expected TARGET to be non-reserved")
	}
}
