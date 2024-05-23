package internal

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/prathoss/cards/pkg"
)

func TestParseCount(t *testing.T) {
	tests := []struct {
		name          string
		param         string
		expectedCount int
		expectedError bool
	}{
		{
			name:          "EmptyParameter",
			param:         "",
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:          "ValidParameter",
			param:         "10",
			expectedCount: 10,
			expectedError: false,
		},
		{
			name:          "NonIntParameter",
			param:         "not-a-number",
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:          "NegativeParameter",
			param:         "-5",
			expectedCount: -5,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?count="+tt.param, nil)
			gotCount, gotInvalidParams := parseCount(req)

			if gotCount != tt.expectedCount {
				t.Errorf("parseCount() gotCount = %v, expectedCount = %v", gotCount, tt.expectedCount)
			}

			if (len(gotInvalidParams) > 0) != tt.expectedError {
				t.Errorf("parseCount() gotInvalidParams = %v, expectedError = %v", gotInvalidParams, tt.expectedError)
			}
		})
	}
}

func TestParseCards(t *testing.T) {
	tests := []struct {
		desc           string
		reqURL         string
		expectedCards  []string
		expectedErrors []pkg.InvalidParam
	}{
		{
			desc:           "Valid card set",
			reqURL:         "/?cards=KH,KD,KS,KC",
			expectedCards:  []string{"KH", "KD", "KS", "KC"},
			expectedErrors: nil,
		},
		{
			desc:          "Invalid card set",
			reqURL:        "/?cards=KH,XD,KS,YC",
			expectedCards: []string{"KH", "XD", "KS", "YC"},
			expectedErrors: []pkg.InvalidParam{
				{
					Name:   "cards",
					Reason: "unrecognised card: XD",
				},
				{
					Name:   "cards",
					Reason: "unrecognised card: YC",
				},
			},
		},
		{
			desc:           "Empty card parameter",
			reqURL:         "/?cards=",
			expectedCards:  []string{},
			expectedErrors: nil,
		},
		{
			desc:           "No card parameter",
			reqURL:         "/",
			expectedCards:  []string{},
			expectedErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.reqURL, nil)
			cards, errors := parseCards(req)

			if !slices.Equal(cards, tt.expectedCards) {
				t.Errorf("Expected cards %v, but got %v", tt.expectedCards, cards)
			}
			if !slices.Equal(errors, tt.expectedErrors) {
				t.Errorf("Expected errors to be %v, but got %v", tt.expectedErrors, errors)
			}
		})
	}
}
