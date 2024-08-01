// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releasesjson

import "testing"

func TestDetermineArchiveURL(t *testing.T) {
	tests := []struct {
		name       string
		archiveURL string
		baseURL    string
		want       string
	}{
		{
			name:       "with custom base URL + path",
			archiveURL: "https://releases.hashicorp.com/terraform/1.8.2/terraform_1.8.2_darwin_amd64.zip",
			baseURL:    "https://myartifactory.company.com/artifactory/hashicorp-remote",
			want:       "https://myartifactory.company.com/artifactory/hashicorp-remote/terraform/1.8.2/terraform_1.8.2_darwin_amd64.zip",
		},
		{
			name:       "with custom base URL + port + path",
			archiveURL: "https://releases.hashicorp.com/terraform/1.8.2/terraform_1.8.2_darwin_amd64.zip",
			baseURL:    "https://myartifactory.company.com:443/artifactory/hashicorp-remote",
			want:       "https://myartifactory.company.com:443/artifactory/hashicorp-remote/terraform/1.8.2/terraform_1.8.2_darwin_amd64.zip",
		},
		{
			name:       "without custom base URL",
			archiveURL: "https://releases.hashicorp.com/terraform/1.8.2/terraform_1.8.2_darwin_amd64.zip",
			baseURL:    "",
			want:       "https://releases.hashicorp.com/terraform/1.8.2/terraform_1.8.2_darwin_amd64.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineArchiveURL(tt.archiveURL, tt.baseURL)
			if err != nil {
				t.Errorf("determineArchiveURL() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("determineArchiveURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
