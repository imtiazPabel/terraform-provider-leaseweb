package shared

import (
	"encoding/json"
)

type ErrorResponse struct {
	CorrelationId string              `json:"correlationId"`
	ErrorCode     string              `json:"errorCode"`
	ErrorMessage  string              `json:"errorMessage"`
	ErrorDetails  map[string][]string `json:"errorDetails"`
}

func NewErrorResponse(jsonStr string) (*ErrorResponse, error) {
	errorResponse := ErrorResponse{}

	err := json.Unmarshal([]byte(jsonStr), &errorResponse)
	if err != nil {
		return nil, err
	}

	return &errorResponse, nil
}
