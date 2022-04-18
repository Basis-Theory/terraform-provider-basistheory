package provider

import (
	"github.com/Basis-Theory/basistheory-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type genericAPIError interface {
	Error() string
	Body() []byte
	Model() interface{}
}

func apiErrorDiagnostics(message string, err error) diag.Diagnostics {
	var errorArgs []interface{}

	if apiError, ok := err.(genericAPIError); ok {
		switch apiError.Model().(type) {
		case basistheory.ValidationProblemDetails:
			message, errorArgs = processValidationProblemDetails(apiError.Model().(basistheory.ValidationProblemDetails), message, errorArgs)
			break
		case basistheory.ProblemDetails:
			message, errorArgs = processProblemDetails(apiError.Model().(basistheory.ProblemDetails), message, errorArgs)
			break
		}
	}

	return diag.Errorf(message, errorArgs...)
}

func processValidationProblemDetails(details basistheory.ValidationProblemDetails, message string, errorArgs []interface{}) (string, []interface{}) {
	addErrorStatus(details.Status, &message, &errorArgs)
	addErrorTitle(details.Title, &message, &errorArgs)
	addErrorDetail(details.Detail, &message, &errorArgs)
	addErrorValidationErrors(details.Errors, &message, &errorArgs)

	return message, errorArgs
}

func processProblemDetails(details basistheory.ProblemDetails, message string, errorArgs []interface{}) (string, []interface{}) {
	addErrorStatus(details.Status, &message, &errorArgs)
	addErrorTitle(details.Title, &message, &errorArgs)
	addErrorDetail(details.Detail, &message, &errorArgs)

	return message, errorArgs
}

func addErrorStatus(status basistheory.NullableInt32, message *string, errorArgs *[]interface{}) {
	if status.IsSet() {
		*message += "\n\tStatus Code: %d"

		*errorArgs = append(*errorArgs, *status.Get())
	}
}

func addErrorTitle(title basistheory.NullableString, message *string, errorArgs *[]interface{}) {
	if title.IsSet() {
		*message += "\n\tTitle: %s"

		*errorArgs = append(*errorArgs, *title.Get())
	}
}

func addErrorDetail(detail basistheory.NullableString, message *string, errorArgs *[]interface{}) {
	if detail.IsSet() {
		*message += "\n\tDetail: %s"

		*errorArgs = append(*errorArgs, *detail.Get())
	}
}

func addErrorValidationErrors(validationErrors map[string][]string, message *string, errorArgs *[]interface{}) {
	if len(validationErrors) == 0 {
		return
	}
	*message += "\n\tErrors:"

	for propertyName, propertyErrors := range validationErrors {
		*message += "\n\t\t%s: %+v"
		*errorArgs = append(*errorArgs, propertyName)
		*errorArgs = append(*errorArgs, propertyErrors)
	}
}
