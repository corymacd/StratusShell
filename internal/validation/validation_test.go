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

func TestValidateNpmPackage(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		wantErr     bool
	}{
		{
			name:        "valid simple package",
			packageName: "express",
			wantErr:     false,
		},
		{
			name:        "valid scoped package",
			packageName: "@playwright/mcp",
			wantErr:     false,
		},
		{
			name:        "valid package with dash",
			packageName: "github-mcp-server",
			wantErr:     false,
		},
		{
			name:        "valid package with dots",
			packageName: "some.package.name",
			wantErr:     false,
		},
		{
			name:        "valid scoped with dash",
			packageName: "@mseep/linear-mcp",
			wantErr:     false,
		},
		{
			name:        "empty package name",
			packageName: "",
			wantErr:     true,
		},
		{
			name:        "package with special chars",
			packageName: "pack@ge",
			wantErr:     true,
		},
		{
			name:        "package with spaces",
			packageName: "my package",
			wantErr:     true,
		},
		{
			name:        "package with semicolon",
			packageName: "package;rm -rf",
			wantErr:     true,
		},
		{
			name:        "package with pipe",
			packageName: "package|ls",
			wantErr:     true,
		},
		{
			name:        "package with ampersand",
			packageName: "package&& rm",
			wantErr:     true,
		},
		{
			name:        "package too long",
			packageName: "a123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNpmPackage(tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNpmPackage(%q) error = %v, wantErr %v", tt.packageName, err, tt.wantErr)
			}
		})
	}
}
