package provider

import (
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk"
	basistheorycore "github.com/Basis-Theory/go-sdk/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"reflect"
)

func apiErrorDiagnostics(message string, err error) diag.Diagnostics {
	var errorArgs []interface{}

	switch e := err.(type) {
	case *basistheory.BadRequestError:
		message, errorArgs = processValidationProblemDetails(e.Body, message, errorArgs)
	case *basistheory.ConflictError:
		message, errorArgs = processProblemDetails(e.Body, message, errorArgs)
	case *basistheory.ForbiddenError:
		message, errorArgs = processProblemDetails(e.Body, message, errorArgs)
	case *basistheory.UnauthorizedError:
		message, errorArgs = processProblemDetails(e.Body, message, errorArgs)
	case *basistheory.UnprocessableEntityError:
		message, errorArgs = processProblemDetails(e.Body, message, errorArgs)
	case *basistheorycore.APIError:
		message, errorArgs = processApiError(*e, message, errorArgs)
	default:
		message, errorArgs = unknownError(message, err, errorArgs)
	}

	return diag.Errorf(message, errorArgs...)
}

func unknownError(message string, err error, errorArgs []interface{}) (string, []interface{}) {
	if err == nil {
		message += "\n\tUnknown Error: (unavailable)"
		return message, errorArgs
	}
	message += "\n\tUnknown error:" + fmt.Sprintf("%s (%s)", err.Error(), reflect.TypeOf(err).String())
	return message, errorArgs
}

func processApiError(err basistheorycore.APIError, message string, errorArgs []interface{}) (string, []interface{}) {
	message += "\n\tStatus Code: %s"
	errorArgs = append(errorArgs, err.StatusCode)
	return message, errorArgs
}

func processValidationProblemDetails(details *basistheory.ValidationProblemDetails, message string, errorArgs []interface{}) (string, []interface{}) {
	if details == nil {
		return message, errorArgs
	}
	addErrorStatus(details.Status, &message, &errorArgs)
	addErrorTitle(details.Title, &message, &errorArgs)
	addErrorDetail(details.Detail, &message, &errorArgs)
	addErrorValidationErrors(details.Errors, &message, &errorArgs)

	return message, errorArgs
}

func processProblemDetails(details *basistheory.ProblemDetails, message string, errorArgs []interface{}) (string, []interface{}) {
	if details == nil {
		return message, errorArgs
	}
	addErrorStatus(details.Status, &message, &errorArgs)
	addErrorTitle(details.Title, &message, &errorArgs)
	addErrorDetail(details.Detail, &message, &errorArgs)

	return message, errorArgs
}

func addErrorStatus(status *int, message *string, errorArgs *[]interface{}) {
	*message += "\n\tStatus Code: %d"
	*errorArgs = append(*errorArgs, *status)
}

func addErrorTitle(title *string, message *string, errorArgs *[]interface{}) {
	*message += "\n\tTitle: %s"
	*errorArgs = append(*errorArgs, *title)
}

func addErrorDetail(detail *string, message *string, errorArgs *[]interface{}) {
	*message += "\n\tDetail: %s"
	*errorArgs = append(*errorArgs, *detail)
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
