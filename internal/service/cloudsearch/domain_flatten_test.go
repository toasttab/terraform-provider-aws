// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	tfcloudsearch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudsearch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// testTime returns a fixed time for testing purposes
func testTime() time.Time {
	return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestFlattenIndexFieldStatuses_PendingDeletion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    []types.IndexFieldStatus
		expected int // expected number of fields in result
		wantErr  bool
	}{
		{
			name: "skip field with PendingDeletion true",
			input: []types.IndexFieldStatus{
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field1"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate:    aws.Time(testTime()),
						State:           types.OptionStateRequiresIndexDocuments,
						UpdateDate:      aws.Time(testTime()),
						PendingDeletion: aws.Bool(true), // Should be skipped
					},
				},
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field2"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate:    aws.Time(testTime()),
						State:           types.OptionStateActive,
						UpdateDate:      aws.Time(testTime()),
						PendingDeletion: aws.Bool(false),
					},
				},
			},
			expected: 1, // Only field2 should be returned
			wantErr:  false,
		},
		{
			name: "skip field with nil Options",
			input: []types.IndexFieldStatus{
				{
					Options: nil, // Should be skipped
					Status: &types.OptionStatus{
						CreationDate: aws.Time(testTime()),
						State:        types.OptionStateActive,
						UpdateDate:   aws.Time(testTime()),
					},
				},
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field2"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate: aws.Time(testTime()),
						State:        types.OptionStateActive,
						UpdateDate:   aws.Time(testTime()),
					},
				},
			},
			expected: 1, // Only field2 should be returned
			wantErr:  false,
		},
		{
			name: "skip field with nil Status",
			input: []types.IndexFieldStatus{
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field1"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: nil, // Should be skipped
				},
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field2"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate: aws.Time(testTime()),
						State:        types.OptionStateActive,
						UpdateDate:   aws.Time(testTime()),
					},
				},
			},
			expected: 1, // Only field2 should be returned
			wantErr:  false,
		},
		{
			name: "all fields pending deletion returns empty",
			input: []types.IndexFieldStatus{
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field1"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate:    aws.Time(testTime()),
						State:           types.OptionStateRequiresIndexDocuments,
						UpdateDate:      aws.Time(testTime()),
						PendingDeletion: aws.Bool(true),
					},
				},
			},
			expected: 0, // All fields should be skipped
			wantErr:  false,
		},
		{
			name: "no pending deletions returns all fields",
			input: []types.IndexFieldStatus{
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field1"),
						IndexFieldType: types.IndexFieldTypeLiteral,
						LiteralOptions: &types.LiteralOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate: aws.Time(testTime()),
						State:        types.OptionStateActive,
						UpdateDate:   aws.Time(testTime()),
					},
				},
				{
					Options: &types.IndexField{
						IndexFieldName: aws.String("field2"),
						IndexFieldType: types.IndexFieldTypeInt,
						IntOptions: &types.IntOptions{
							ReturnEnabled: aws.Bool(true),
						},
					},
					Status: &types.OptionStatus{
						CreationDate: aws.Time(testTime()),
						State:        types.OptionStateActive,
						UpdateDate:   aws.Time(testTime()),
					},
				},
			},
			expected: 2, // Both fields should be returned
			wantErr:  false,
		},
		{
			name:     "empty input returns nil",
			input:    []types.IndexFieldStatus{},
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfcloudsearch.FlattenIndexFieldStatuses(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if len(result) != tc.expected {
				t.Errorf("expected %d fields, got %d", tc.expected, len(result))
			}

			// Verify no nil entries in result (which would cause panic in TypeSet)
			for i, item := range result {
				if item == nil {
					t.Errorf("result[%d] is nil, which would cause panic in TypeSet", i)
				}
			}
		})
	}
}

func TestFlattenIndexFieldStatus_PendingDeletion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		input     types.IndexFieldStatus
		expectNil bool
		wantErr   bool
	}{
		{
			name: "pending deletion returns nil",
			input: types.IndexFieldStatus{
				Options: &types.IndexField{
					IndexFieldName: aws.String("test_field"),
					IndexFieldType: types.IndexFieldTypeLiteral,
					LiteralOptions: &types.LiteralOptions{
						ReturnEnabled: aws.Bool(true),
					},
				},
				Status: &types.OptionStatus{
					CreationDate:    aws.Time(testTime()),
					State:           types.OptionStateRequiresIndexDocuments,
					UpdateDate:      aws.Time(testTime()),
					PendingDeletion: aws.Bool(true),
				},
			},
			expectNil: true,
			wantErr:   false,
		},
		{
			name: "nil Options returns nil",
			input: types.IndexFieldStatus{
				Options: nil,
				Status: &types.OptionStatus{
					CreationDate: aws.Time(testTime()),
					State:        types.OptionStateActive,
					UpdateDate:   aws.Time(testTime()),
				},
			},
			expectNil: true,
			wantErr:   false,
		},
		{
			name: "nil Status returns nil",
			input: types.IndexFieldStatus{
				Options: &types.IndexField{
					IndexFieldName: aws.String("test_field"),
					IndexFieldType: types.IndexFieldTypeLiteral,
					LiteralOptions: &types.LiteralOptions{
						ReturnEnabled: aws.Bool(true),
					},
				},
				Status: nil,
			},
			expectNil: true,
			wantErr:   false,
		},
		{
			name: "active field returns data",
			input: types.IndexFieldStatus{
				Options: &types.IndexField{
					IndexFieldName: aws.String("test_field"),
					IndexFieldType: types.IndexFieldTypeLiteral,
					LiteralOptions: &types.LiteralOptions{
						ReturnEnabled: aws.Bool(true),
					},
				},
				Status: &types.OptionStatus{
					CreationDate:    aws.Time(testTime()),
					State:           types.OptionStateActive,
					UpdateDate:      aws.Time(testTime()),
					PendingDeletion: aws.Bool(false),
				},
			},
			expectNil: false,
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfcloudsearch.FlattenIndexFieldStatus(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if tc.expectNil {
				if result != nil {
					t.Errorf("expected nil result, got: %v", result)
				}
			} else {
				if result == nil {
					t.Error("expected non-nil result, got nil")
				}
				// Verify expected fields are present
				if name, ok := result[names.AttrName]; !ok || name == "" {
					t.Errorf("expected name field in result, got: %v", result)
				}
			}
		})
	}
}
