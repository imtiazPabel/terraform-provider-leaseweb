package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("adds generic error when err is nil", func(t *testing.T) {
		diags := diag.Diagnostics{}
		httpResponse := http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(bytes.NewReader([]byte(``))),
		}

		Error(context.TODO(), &diags, "summary", nil, &httpResponse)

		assert.Len(t, diags.Errors(), 1)
		assert.Equal(t, "summary", diags.Errors()[0].Summary())
		assert.Equal(t, "An error has occurred in the program. Please consider opening an issue.", diags.Errors()[0].Detail())
	})

	t.Run("handles SDK error when HTTP status is non-error", func(t *testing.T) {
		diags := diag.Diagnostics{}
		httpResponse := http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
		}

		Error(context.TODO(), &diags, "summary", errors.New("SDK error: enum doesn't match."), &httpResponse)

		assert.Len(t, diags.Errors(), 1)
		assert.Equal(t, "summary", diags.Errors()[0].Summary())
		assert.Equal(t, "SDK error: enum doesn't match.", diags.Errors()[0].Detail())
	})

	t.Run("adds attribute error from HTTP validation details", func(t *testing.T) {
		diags := diag.Diagnostics{}
		httpResponse := http.Response{
			StatusCode: 500,
			Body: io.NopCloser(
				bytes.NewReader(
					[]byte(`
            {
              "correlationId": "correlationId",
              "errorCode": "errorCode",
              "errorMessage": "errorMessage",
              "errorDetails":  {
                "name": ["the name is invalid"]
              }
            }
          `),
				),
			),
		}

		Error(context.TODO(), &diags, "summary", errors.New("error content"), &httpResponse)

		attributePath := path.Root("name")
		want := diag.Diagnostics{}
		want.AddAttributeError(attributePath, "summary", "the name is invalid")
		assert.Equal(t, want, diags.Errors())
	})

	t.Run("handles regular HTTP error response", func(t *testing.T) {
		diags := diag.Diagnostics{}
		httpResponse := http.Response{
			StatusCode: 500,
			Body: io.NopCloser(
				bytes.NewReader(
					[]byte(`
            {
              "correlationId": "correlationId",
              "errorCode": "404",
              "errorMessage": "Server not found"
            }
          `),
				),
			),
		}

		Error(context.TODO(), &diags, "summary", errors.New("error content"), &httpResponse)

		assert.Len(t, diags.Errors(), 1)
		assert.Equal(t, "summary", diags.Errors()[0].Summary())
		assert.Equal(t, "{\n  \"errorCode\": \"404\",\n  \"errorMessage\": \"Server not found\",\n  \"correlationId\": \"correlationId\"\n}", diags.Errors()[0].Detail())
	})
}

func ExampleError() {
	diags := diag.Diagnostics{}

	httpResponse := http.Response{
		StatusCode: 500,
		Body: io.NopCloser(
			bytes.NewReader(
				[]byte(`
{
  "correlationId": "correlationId",
  "errorCode": "errorCode",
  "errorMessage": "errorMessage",
  "errorDetails":  {
  "name": ["the name is invalid"]
  }
}
          `),
			),
		),
	}

	Error(context.TODO(), &diags, "summary", errors.New("error content"), &httpResponse)

	fmt.Println(diags.Errors())
	// Output: [{{the name is invalid summary} {[name]}}]
}

func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		attr      []string
		expected  string
	}{
		{
			name:      "test_resource",
			operation: "Creating resource",
			attr:      []string{"username", "testuser"},
			expected:  "Creating resource test_resource for username: \"testuser\"",
		},
		{
			name:      "test_data_source",
			operation: "Reading data",
			attr:      []string{"unknown-key-without-value"},
			expected:  "Reading data test_data_source",
		},
		{
			name:      "test_data_source",
			operation: "Reading data",
			attr:      []string{}, // No attributes
			expected:  "Reading data test_data_source",
		},
		{
			name:      "test_resource",
			operation: "Updating resource",
			attr:      []string{"username", "newuser", "email", "newuser@example.com"},
			expected:  "Updating resource test_resource for username: \"newuser\" email: \"newuser@example.com\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSummary(tt.name, tt.operation, tt.attr...)
			if got != tt.expected {
				t.Errorf("BuildSummary() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func ExampleBuildSummary() {
	summary := BuildSummary("test_resource", "Updating resource", "username", "newuser", "email", "newuser@example.com")
	fmt.Println(summary)
	// Output: Updating resource test_resource for username: "newuser" email: "newuser@example.com"
}
