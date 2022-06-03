package provider

import (
	"fmt"
	"github.com/Basis-Theory/basistheory-go/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"net/http"
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

func TestErrorUtils_apiErrorDiagnostics_shouldAddValidationProblemDetailsToError(t *testing.T) {
	var expectedStatusCode int32 = 400
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"
	expectedErrors := map[string][]string{
		"prop1": {"error1"},
	}
	statusCode := basistheory.NullableInt32{}
	statusCode.Set(&expectedStatusCode)

	title := basistheory.NullableString{}
	title.Set(&expectedTitle)

	detail := basistheory.NullableString{}
	detail.Set(&expectedDetail)

	var validationProblemDetails interface{}
	validationProblemDetails = basistheory.ValidationProblemDetails{
		Status: statusCode,
		Title:  title,
		Detail: detail,
		Errors: expectedErrors,
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s\n\tErrors:\n\t\t%s: %+v", originalMessage, expectedStatusCode, expectedTitle, expectedDetail, "prop1", expectedErrors["prop1"])

	var apiError genericAPIError
	apiError = testGenericOpenAPIError{
		body:  nil,
		error: "",
		model: validationProblemDetails,
	}

	actual := apiErrorDiagnostics(originalMessage, nil, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldAddProblemDetailsToError(t *testing.T) {
	var expectedStatusCode int32 = 400
	expectedTitle := "This is my title"
	expectedDetail := "These are my details"
	statusCode := basistheory.NullableInt32{}
	statusCode.Set(&expectedStatusCode)

	title := basistheory.NullableString{}
	title.Set(&expectedTitle)

	detail := basistheory.NullableString{}
	detail.Set(&expectedDetail)

	problemDetails := basistheory.ProblemDetails{
		Status: statusCode,
		Title:  title,
		Detail: detail,
	}

	originalMessage := "Error encountered"

	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %d\n\tTitle: %s\n\tDetail: %s", originalMessage, expectedStatusCode, expectedTitle, expectedDetail)

	var apiError genericAPIError
	apiError = testGenericOpenAPIError{
		body:  nil,
		error: "",
		model: problemDetails,
	}

	actual := apiErrorDiagnostics(originalMessage, nil, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleEmptyValidationProblemDetails(t *testing.T) {
	expected := "Error encountered"

	var apiError genericAPIError
	apiError = testGenericOpenAPIError{
		body:  nil,
		error: "",
		model: basistheory.ValidationProblemDetails{},
	}

	actual := apiErrorDiagnostics(expected, nil, apiError)

	assert.Equal(t, expected, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleEmptyProblemDetails(t *testing.T) {
	expected := "Error encountered"

	var apiError genericAPIError
	apiError = testGenericOpenAPIError{
		body:  nil,
		error: "",
		model: basistheory.ProblemDetails{},
	}

	actual := apiErrorDiagnostics(expected, nil, apiError)

	assert.Equal(t, expected, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleUnknownErrorModel(t *testing.T) {
	originalMessage := "Error encountered"

	var apiError genericAPIError
	apiError = testGenericOpenAPIError{
		body:  nil,
		error: "",
		model: testGenericOpenAPIError{
			body:  nil,
			error: "foo",
			model: nil,
		},
	}

	response := http.Response{Status: "500"}
	expectedErrorMessage := fmt.Sprintf("%s\n\tStatus Code: %s", originalMessage, response.Status)

	actual := apiErrorDiagnostics(originalMessage, &response, apiError)

	assert.Equal(t, expectedErrorMessage, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}

func TestErrorUtils_apiErrorDiagnostics_shouldHandleNilError(t *testing.T) {
	expected := "Error encountered"

	actual := apiErrorDiagnostics(expected, nil, nil)

	assert.Equal(t, expected, actual[0].Summary)
	assert.Equal(t, diag.Error, actual[0].Severity)
}
