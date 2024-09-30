package provider

import (
	basistheory "github.com/Basis-Theory/go-sdk"
	basistheorycore "github.com/Basis-Theory/go-sdk/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func apiErrorDiagnosticsV2(message string, err error) diag.Diagnostics {
	var errorArgs []interface{}

	switch e := err.(type) {
	case basistheory.BadRequestError:
		message, errorArgs = processValidationProblemDetailsV2(e.Body, message, errorArgs)
	case basistheory.ForbiddenError:
	case basistheory.UnauthorizedError:
	case basistheory.UnprocessableEntityError:
		message, errorArgs = processProblemDetailsV2(e.Body, message, errorArgs)
	case *basistheorycore.APIError:
		message, errorArgs = processApiError(*e, message, errorArgs)
	default:
		message, errorArgs = unknownError(err, message, errorArgs)
	}

	return diag.Errorf(message, errorArgs...)
}

func unknownError(err error, message string, errorArgs []interface{}) (string, []interface{}) {
	message += "\n\tUnknown error"
	return message, errorArgs
}

func processApiError(err basistheorycore.APIError, message string, errorArgs []interface{}) (string, []interface{}) {
	message += "\n\tStatus Code: %s"
	errorArgs = append(errorArgs, err.StatusCode)
	return message, errorArgs
}

func processValidationProblemDetailsV2(details *basistheory.ValidationProblemDetails, message string, errorArgs []interface{}) (string, []interface{}) {
	addErrorStatusV2(details.Status, &message, &errorArgs)
	addErrorTitleV2(details.Title, &message, &errorArgs)
	addErrorDetailV2(details.Detail, &message, &errorArgs)
	addErrorValidationErrorsV2(details.Errors, &message, &errorArgs)

	return message, errorArgs
}

func processProblemDetailsV2(details *basistheory.ProblemDetails, message string, errorArgs []interface{}) (string, []interface{}) {
	addErrorStatusV2(details.Status, &message, &errorArgs)
	addErrorTitleV2(details.Title, &message, &errorArgs)
	addErrorDetailV2(details.Detail, &message, &errorArgs)

	return message, errorArgs
}

func addErrorStatusV2(status *int, message *string, errorArgs *[]interface{}) {
	*message += "\n\tStatus Code: %d"
	*errorArgs = append(*errorArgs, *status)
}

func addErrorTitleV2(title *string, message *string, errorArgs *[]interface{}) {
	*message += "\n\tTitle: %s"
	*errorArgs = append(*errorArgs, *title)
}

func addErrorDetailV2(detail *string, message *string, errorArgs *[]interface{}) {
	*message += "\n\tDetail: %s"
	*errorArgs = append(*errorArgs, *detail)
}

func addErrorValidationErrorsV2(validationErrors map[string][]string, message *string, errorArgs *[]interface{}) {
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
