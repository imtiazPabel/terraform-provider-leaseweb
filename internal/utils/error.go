package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const DefaultErrMsg = "An error has occurred in the program. Please consider opening an issue."

type Error struct {
	err  error
	resp *http.Response
}

// TODO: we need to merge error.go and log_error.go and have unified error/logging functionality.
func (e Error) Error() string {
	// Check if the response or its body is nil,
	//or if the status code is not an error.
	if e.resp == nil || e.resp.Body == nil || e.resp.StatusCode < 400 {
		return e.err.Error()
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}(e.resp.Body)

	// Try to decode the response body as JSON.
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(e.resp.Body).Decode(&errorResponse); err != nil {
		return e.err.Error()
	}

	if msg := buildErrorMessage(errorResponse); msg != "" && msg != "{}" {
		return msg
	}

	// Default to the original error message if no relevant information is found.
	return e.err.Error()
}

// Helper function to build an error message from the decoded JSON.
func buildErrorMessage(errorResponse map[string]interface{}) string {
	// Create a map to store only the fields we care about.
	output := make(map[string]interface{})

	if errorCode, ok := errorResponse["errorCode"]; ok {
		output["errorCode"] = errorCode
	}
	if errorMessage, ok := errorResponse["errorMessage"]; ok {
		output["errorMessage"] = errorMessage
	}
	if userMessage, ok := errorResponse["userMessage"]; ok {
		output["userMessage"] = userMessage
	}
	if correlationId, ok := errorResponse["correlationId"]; ok {
		output["correlationId"] = correlationId
	}
	if errorDetails, ok := errorResponse["errorDetails"]; ok {
		output["errorDetails"] = errorDetails
	}

	// Encode the output map as a JSON string.
	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonOutput)
}

func NewError(resp *http.Response, err error) Error {
	return Error{
		resp: resp,
		err:  err,
	}
}

// normalizeErrorResponseKey converts an api key path to a string
// that HandleSdkError can handle.
// `instanceId` & `instance.id` both become `instance_id`.
func normalizeErrorResponseKey(key string) string {
	// Assume that the key has the format `contract.id`
	//if any dots are found.
	if strings.Contains(key, ".") {
		return strings.ToLower(strings.Replace(key, ".", "_", -1))
	}

	// If no dots are found, assume camel case.
	m := regexp.MustCompile("[A-Z]")
	res := m.ReplaceAllStringFunc(key, func(s string) string {
		return "_" + s
	})

	return strings.ToLower(res)
}

// HandleSdkError takes a server response & error
// and maps errors to the appropriate attributes.
// If an attribute cannot be found,
// the error is shown to the user on a resource level.
// A DEBUG log is also created with all the relevant information.
func HandleSdkError(
	summary string,
	httpResponse *http.Response,
	err error,
	diags *diag.Diagnostics,
	ctx context.Context,
) {
	// Nothing to do when httpResponse does not exist.
	if httpResponse == nil {
		handleError(summary, err, diags)
		return
	}

	response, err := newResponse(httpResponse.Body)
	if err != nil {
		handleError(summary, nil, diags)
		return
	}

	// Create DEBUG log with httpResponse body.
	response.newDebugLog(ctx, summary)

	// Convert httpResponse buffer to ErrorResponse object.
	err = response.setErrors(summary, diags)
	if err != nil {
		handleError(summary, err, diags)
		return
	}
}

type ErrorResponse struct {
	CorrelationId string              `json:"correlationId,omitempty"`
	ErrorCode     string              `json:"errorCode,omitempty"`
	ErrorMessage  string              `json:"errorMessage,omitempty"`
	ErrorDetails  map[string][]string `json:"errorDetails,omitempty"`
}

// If passed, add the specific error to diags. If no error is passed then show the default error.
func handleError(
	summary string,
	err error,
	diags *diag.Diagnostics,
) {
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}

	diags.AddError(summary, DefaultErrMsg)
}

type response struct {
	buf           *strings.Builder
	errorResponse ErrorResponse
	responseMap   map[string]interface{}
}

// If set show the responseMap. If the responseMap is not set show the httpResponse body.
func (r response) newDebugLog(
	ctx context.Context,
	summary string,
) {
	if r.responseMap == nil {
		tflog.Debug(
			ctx,
			summary,
			map[string]any{"httpResponse": fmt.Sprintf("%v", r.buf.String())},
		)
		return
	}

	tflog.Debug(ctx, summary, map[string]any{"response": r.responseMap})
}

// Try to set attribute errors, if that's not possible set a global resource error.
func (r response) setErrors(summary string, diags *diag.Diagnostics) error {
	errorSet := r.setAttributeErrors(summary, diags)

	if !errorSet {
		err := r.setGlobalError(summary, diags)
		if err != nil {
			return err
		}
	}

	return nil
}

// Return true if an attribute error is set, otherwise return false.
func (r response) setAttributeErrors(
	summary string,
	diags *diag.Diagnostics,
) bool {
	// Convert key returned from api to an attribute path.
	// I.e.: []string{"image", "id"}.
	errorSet := false
	for errorKey, errorDetailList := range r.errorResponse.ErrorDetails {
		normalizedErrorKey := normalizeErrorResponseKey(errorKey)
		mapKeys := strings.Split(normalizedErrorKey, "_")
		attributePath := path.Root(mapKeys[0])

		// Every element in the map goes one level deeper.
		for _, mapKey := range mapKeys[1:] {
			attributePath = attributePath.AtMapKey(mapKey)
		}

		// Each attribute can have multiple errors.
		for _, errorDetail := range errorDetailList {
			diags.AddAttributeError(attributePath, summary, errorDetail)
		}
		errorSet = true
	}

	return errorSet
}

func (r response) setGlobalError(
	summary string,
	diags *diag.Diagnostics,
) error {
	errorResponseString, err := json.MarshalIndent(
		r.errorResponse,
		"",
		" ",
	)
	if err != nil {
		return err
	}

	diags.AddError(summary, string(errorResponseString))
	return nil
}

func newResponse(body io.ReadCloser) (*response, error) {
	var responseMap map[string]interface{}

	// Try to read httpResponse body into buffer
	buf := new(strings.Builder)
	_, err := io.Copy(buf, body)
	if err != nil {
		return nil, err
	}

	errorResponse := ErrorResponse{}

	err = json.Unmarshal([]byte(buf.String()), &errorResponse)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(buf.String()), &responseMap)
	if err != nil {
		responseMap = nil
	}

	return &response{
		buf:           buf,
		errorResponse: errorResponse,
		responseMap:   responseMap,
	}, nil
}
