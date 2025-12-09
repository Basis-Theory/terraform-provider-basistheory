package provider

import (
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk/v4"
	"github.com/Basis-Theory/go-sdk/v4/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testGenericOpenAPIError struct {
	body  []byte
	error string
	model interface{}
}

func (e testGenericOpenAPIError) Error() string {
	return e.error
}

func (e testGenericOpenAPIError) Body() []byte {
	return e.body
}

func (e testGenericOpenAPIError) Model() interface{} {
	return e.model
}

func TestErrorUtils_apiErrorDiagnostics_BadRequestError_shouldAddValidationProblemDetailsToError(t *testing.T) {
	var expectedStatusCode = 400
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"
	expectedErrors := map[string][]string{
		"prop1": {"error1"},
	}

	apiErr := &core.APIError{
		StatusCode: 400,
	}

	var validationProblemDetails = &basistheory.ValidationProblemDetails{
		Status: getIntPointer(expectedStatusCode),
		Title:  getStringPointer(expectedTitle),
		Detail: getStringPointer(expectedDetail),
		Errors: expectedErrors,
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s\n\tErrors:\n\t\t%s: %+v", originalMessage, expectedStatusCode, expectedTitle, expectedDetail, "prop1", expectedErrors["prop1"])

	var apiError = &basistheory.BadRequestError{
		Body:     validationProblemDetails,
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(originalMessage, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_Conflict_shouldAddProblemDetailsToError(t *testing.T) {
	var expectedStatusCode = 409
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"

	apiErr := &core.APIError{
		StatusCode: 400,
	}

	problemDetails := &basistheory.ProblemDetails{
		Status: getIntPointer(expectedStatusCode),
		Title:  getStringPointer(expectedTitle),
		Detail: getStringPointer(expectedDetail),
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s", originalMessage, expectedStatusCode, expectedTitle, expectedDetail)

	var apiError = &basistheory.ConflictError{
		Body:     problemDetails,
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(originalMessage, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_Forbidden_shouldAddProblemDetailsToError(t *testing.T) {
	var expectedStatusCode = 403
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"

	apiErr := &core.APIError{
		StatusCode: 400,
	}

	problemDetails := &basistheory.ProblemDetails{
		Status: getIntPointer(expectedStatusCode),
		Title:  getStringPointer(expectedTitle),
		Detail: getStringPointer(expectedDetail),
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s", originalMessage, expectedStatusCode, expectedTitle, expectedDetail)

	var apiError = &basistheory.ForbiddenError{
		Body:     problemDetails,
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(originalMessage, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_UnauthorizedError_shouldAddProblemDetailsToError(t *testing.T) {
	var expectedStatusCode = 401
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"

	apiErr := &core.APIError{
		StatusCode: 400,
	}

	problemDetails := &basistheory.ProblemDetails{
		Status: getIntPointer(expectedStatusCode),
		Title:  getStringPointer(expectedTitle),
		Detail: getStringPointer(expectedDetail),
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s", originalMessage, expectedStatusCode, expectedTitle, expectedDetail)

	var apiError = &basistheory.UnauthorizedError{
		Body:     problemDetails,
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(originalMessage, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_UnprocessableEntityError_shouldAddProblemDetailsToError(t *testing.T) {
	var expectedStatusCode = 422
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"

	apiErr := &core.APIError{
		StatusCode: 400,
	}

	problemDetails := &basistheory.ProblemDetails{
		Status: getIntPointer(expectedStatusCode),
		Title:  getStringPointer(expectedTitle),
		Detail: getStringPointer(expectedDetail),
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s", originalMessage, expectedStatusCode, expectedTitle, expectedDetail)

	var apiError = &basistheory.UnprocessableEntityError{
		Body:     problemDetails,
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(originalMessage, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleEmptyValidationProblemDetails(t *testing.T) {
	expected := "Error encountered"

	apiErr := &core.APIError{
		StatusCode: 400,
	}

	var apiError = &basistheory.BadRequestError{
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(expected, apiError)

	assert.Equal(t, expected, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleEmptyProblemDetails(t *testing.T) {
	expected := "Error encountered"

	apiErr := &core.APIError{
		StatusCode: 409,
	}

	var apiError = &basistheory.ConflictError{
		APIError: apiErr,
	}

	actual := apiErrorDiagnostics(expected, apiError)

	assert.Equal(t, expected, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleUnknownErrorModel(t *testing.T) {
	originalMessage := "Error encountered"

	var apiError testGenericOpenAPIError
	apiError = testGenericOpenAPIError{
		body:  nil,
		error: "",
		model: testGenericOpenAPIError{
			body:  nil,
			error: "foo",
			model: nil,
		},
	}

	expectedErrorMessage := fmt.Sprintf("%s\n\tUnknown error: (provider.testGenericOpenAPIError)", originalMessage)

	actual := apiErrorDiagnostics(originalMessage, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleNilError(t *testing.T) {
	expected := "Error encountered"

	actual := apiErrorDiagnostics(expected, nil)

	assert.Equal(t, expected+"\n\tUnknown Error: (unavailable)", actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}
