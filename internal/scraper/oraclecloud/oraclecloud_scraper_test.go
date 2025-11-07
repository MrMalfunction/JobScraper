package oraclecloud

import (
	"testing"
)

func TestTransformBrowserURLToAPIURL(t *testing.T) {
	tests := []struct {
		name        string
		browserURL  string
		expectedURL string
		expectError bool
	}{
		{
			name:        "Valid URL with all parameters",
			browserURL:  "https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?lastSelectedFacet=CATEGORIES&location=United+States&locationId=300000000289738&locationLevel=country&mode=location&selectedCategoriesFacet=300000086152753&selectedPostingDatesFacet=7",
			expectedURL: "https://jpmc.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=findReqs;siteNumber=CX_1001;limit=25;offset=0;selectedCategoriesFacet=300000086152753;selectedPostingDatesFacet=7;sortBy=POSTING_DATES_DESC",
			expectError: false,
		},
		{
			name:        "Valid URL without selectedPostingDatesFacet (should default to 7)",
			browserURL:  "https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?lastSelectedFacet=CATEGORIES&selectedCategoriesFacet=300000086152753",
			expectedURL: "https://jpmc.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=findReqs;siteNumber=CX_1001;limit=25;offset=0;selectedCategoriesFacet=300000086152753;selectedPostingDatesFacet=7;sortBy=POSTING_DATES_DESC",
			expectError: false,
		},
		{
			name:        "Valid URL without selectedCategoriesFacet (category is optional)",
			browserURL:  "https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?location=United+States",
			expectedURL: "https://jpmc.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=findReqs;siteNumber=CX_1001;limit=25;offset=0;selectedPostingDatesFacet=7;sortBy=POSTING_DATES_DESC",
			expectError: false,
		},
		{
			name:        "Valid URL with multiple categories (semicolon separated)",
			browserURL:  "https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_1001/jobs?lastSelectedFacet=CATEGORIES&location=United+States&locationId=300000000289738&locationLevel=country&mode=location&selectedCategoriesFacet=300000086249838%3B300000086152753&selectedPostingDatesFacet=7",
			expectedURL: "https://jpmc.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=findReqs;siteNumber=CX_1001;limit=25;offset=0;selectedCategoriesFacet=300000086249838%3B300000086152753;selectedPostingDatesFacet=7;sortBy=POSTING_DATES_DESC",
			expectError: false,
		},
		{
			name:        "Different site number",
			browserURL:  "https://example.fa.oraclecloud.com/hcmUI/CandidateExperience/en/sites/CX_2002/jobs?selectedCategoriesFacet=123456",
			expectedURL: "https://example.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=findReqs;siteNumber=CX_2002;limit=25;offset=0;selectedCategoriesFacet=123456;selectedPostingDatesFacet=7;sortBy=POSTING_DATES_DESC",
			expectError: false,
		},
		{
			name:        "Missing site number in path",
			browserURL:  "https://jpmc.fa.oraclecloud.com/hcmUI/CandidateExperience/en/jobs?selectedCategoriesFacet=300000086152753",
			expectedURL: "",
			expectError: true,
		},
		{
			name:        "Invalid URL",
			browserURL:  "not-a-valid-url",
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TransformBrowserURLToAPIURL(tt.browserURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expectedURL {
					t.Errorf("Expected URL:\n%s\nGot:\n%s", tt.expectedURL, result)
				}
			}
		})
	}
}

func TestParseOracleAPIURL(t *testing.T) {
	tests := []struct {
		name               string
		apiURL             string
		expectedBase       string
		expectedSiteNumber string
		expectError        bool
	}{
		{
			name:               "Valid API URL",
			apiURL:             "https://jpmc.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList&finder=findReqs;siteNumber=CX_1001;limit=25;offset=0;selectedCategoriesFacet=300000086152753;sortBy=POSTING_DATES_DESC",
			expectedBase:       "https://jpmc.fa.oraclecloud.com",
			expectedSiteNumber: "CX_1001",
			expectError:        false,
		},
		{
			name:               "API URL without finder parameter",
			apiURL:             "https://jpmc.fa.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true",
			expectedBase:       "",
			expectedSiteNumber: "",
			expectError:        true,
		},
		{
			name:               "Invalid URL",
			apiURL:             "not-a-valid-url",
			expectedBase:       "",
			expectedSiteNumber: "",
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, siteNumber, err := ParseOracleAPIURL(tt.apiURL)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if baseURL != tt.expectedBase {
					t.Errorf("Expected base URL: %s, got: %s", tt.expectedBase, baseURL)
				}
				if siteNumber != tt.expectedSiteNumber {
					t.Errorf("Expected siteNumber: %s, got: %s", tt.expectedSiteNumber, siteNumber)
				}
			}
		})
	}
}
