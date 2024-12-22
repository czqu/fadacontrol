package utils

import "testing"

func TestSplitWindowsAccount(t *testing.T) {
	tests := []struct {
		name           string
		account        string
		expectedDomain string
		expectedUser   string
	}{
		{
			name:           "Valid account with domain and username",
			account:        "domain\\user",
			expectedDomain: "domain",
			expectedUser:   "user",
		},
		{
			name:           "No domain, only username",
			account:        "user",
			expectedDomain: "",
			expectedUser:   "user",
		},
		{
			name:           "Only domain, no username",
			account:        "domain\\",
			expectedDomain: "domain",
			expectedUser:   "",
		},
		{
			name:           "Empty account string",
			account:        "",
			expectedDomain: "",
			expectedUser:   "",
		},
		{
			name:           "Only backslash",
			account:        "\\",
			expectedDomain: "",
			expectedUser:   "",
		},
		{
			name:           "Multiple backslashes",
			account:        "domain\\subdomain\\user",
			expectedDomain: "domain",
			expectedUser:   "subdomain\\user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, username := SplitWindowsAccount(tt.account)

			if domain != tt.expectedDomain {
				t.Errorf("Expected domain %q, got %q", tt.expectedDomain, domain)
			}

			if username != tt.expectedUser {
				t.Errorf("Expected username %q, got %q", tt.expectedUser, username)
			}
		})
	}
}
