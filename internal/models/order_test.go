package models

import (
	"testing"

	"PWZ1.0/internal/models/domainErrors"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalPrice(t *testing.T) {
	tests := []struct {
		name          string
		pkgType       PackageType
		initialPrice  float32
		expectedPrice float32
	}{
		{"PackageBag", PackageBag, 10, 15},
		{"PackageBox", PackageBox, 0, 20},
		{"PackageTape", PackageTape, 5, 6},
		{"PackageBagTape", PackageBagTape, 1, 7},
		{"PackageBoxTape", PackageBoxTape, 10, 31},
		{"PackageUnspecified", PackageUnspecified, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			order := &Order{
				Price:       tt.initialPrice,
				PackageType: tt.pkgType,
			}
			order.CalculateTotalPrice()
			assert.Equal(t, tt.expectedPrice, order.Price)
		})
	}
}

func TestValidationWeight(t *testing.T) {
	tests := []struct {
		name      string
		pkgType   PackageType
		weight    float32
		wantError error
	}{
		{"BagUnderLimit", PackageBag, 5, nil},
		{"BagOverLimit", PackageBag, 10, domainErrors.ErrWeightTooHeavy},
		{"BoxUnderLimit", PackageBox, 29.9, nil},
		{"BoxOverLimit", PackageBox, 30, domainErrors.ErrWeightTooHeavy},
		{"TapeAnyWeight", PackageTape, 1000, nil},
		{"Unspecified", PackageUnspecified, 1000, nil},
		{"InvalidPackage", PackageType("мешок с дыркой"), 1, domainErrors.ErrInvalidPackage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			order := &Order{
				PackageType: tt.pkgType,
				Weight:      tt.weight,
			}
			err := order.ValidationWeight()
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestParseActionType(t *testing.T) {
	tests := []struct {
		input    string
		expected ActionType
	}{
		{"issue", ActionTypeIssue},
		{"return", ActionTypeReturn},
		{"unknown", ActionTypeUnspecified},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := ParseActionType(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestActionType_String(t *testing.T) {
	tests := []struct {
		action   ActionType
		expected string
	}{
		{ActionTypeIssue, "issue"},
		{ActionTypeReturn, "return"},
		{ActionTypeUnspecified, "unspecified"},
		{ActionType(100500), "unspecified"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			got := tt.action.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}
