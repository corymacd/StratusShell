package validation

import (
	"testing"
)

func TestValidateGroupname(t *testing.T) {
	tests := []struct {
		name      string
		groupname string
		wantErr   bool
	}{
		{
			name:      "valid docker group",
			groupname: "docker",
			wantErr:   false,
		},
		{
			name:      "valid group with underscore prefix",
			groupname: "_testgroup",
			wantErr:   false,
		},
		{
			name:      "valid group with dash",
			groupname: "test-group",
			wantErr:   false,
		},
		{
			name:      "valid group with underscore",
			groupname: "test_group",
			wantErr:   false,
		},
		{
			name:      "empty groupname",
			groupname: "",
			wantErr:   true,
		},
		{
			name:      "groupname with special chars",
			groupname: "test@group",
			wantErr:   true,
		},
		{
			name:      "groupname with uppercase",
			groupname: "TestGroup",
			wantErr:   true,
		},
		{
			name:      "groupname with path traversal",
			groupname: "../shadow",
			wantErr:   true,
		},
		{
			name:      "reserved groupname root",
			groupname: "root",
			wantErr:   true,
		},
		{
			name:      "reserved groupname sudo",
			groupname: "sudo",
			wantErr:   true,
		},
		{
			name:      "groupname too long",
			groupname: "abcdefghijklmnopqrstuvwxyz1234567",
			wantErr:   true,
		},
		{
			name:      "groupname starting with number",
			groupname: "1testgroup",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupname(tt.groupname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGroupname(%q) error = %v, wantErr %v", tt.groupname, err, tt.wantErr)
			}
		})
	}
}
