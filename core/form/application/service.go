/*
This package belongs to the form module.

The Types defined in this application packages are meant to be used in your controller to process forms
*/
package application

import (
	"context"
	"net/url"
	"reflect"
	"strings"

	"flamingo.me/flamingo/core/form/domain"
	"flamingo.me/flamingo/framework/web"
	"gopkg.in/go-playground/validator.v9"
)

/*
ProcessFormRequest: Parses and Validates a Request to a Form - with the Help of the passed FormService

It calls the ParseFormData() method of the passed formService.
Also, in case the form was submitted, it calls ValidateFormData() method of the formService (passing the parsed data)

Validation is only called if request was send via "POST".
Also you can skip validation by passing a "novalidate" parameter. (In this cases form.IsSubmitted stays false)

*/
func ProcessFormRequest(ctx context.Context, r *web.Request, formService domain.FormService) (domain.Form, error) {
	form := domain.Form{}

	urlValues, err := getPostValues(ctx, r)
	if err != nil {
		form.ValidationInfo.AddGeneralUnknownError(err)
		return form, err
	}
	form.OriginalPostValues = urlValues

	form.Data, err = parseFormData(ctx, r, urlValues, formService)
	if err != nil {
		form.ValidationInfo.AddGeneralUnknownError(err)
		return form, err
	}

	form.ValidationRules = extractValidationRules(form)

	//Run Validation only if form was submitted
	if urlValues.Get("novalidate") != "true" && r.Request().Method == "POST" {
		form.IsSubmitted = true
		err = nil
		if validatingFormService, ok := formService.(domain.ValidateFormData); ok {
			form.ValidationInfo, err = validatingFormService.ValidateFormData(form.Data)
		} else if validatingFormServiceWithContext, ok := formService.(domain.ValidateFormDataWithContext); ok {
			form.ValidationInfo, err = validatingFormServiceWithContext.ValidateFormDataWithContext(ctx, form.Data)
		}
		if err != nil {
			form.ValidationInfo = ValidationErrorsToValidationInfo(err)
		}
	} else {
		if getDefaultFormDataType, ok := formService.(domain.GetDefaultFormData); ok {
			form.Data = getDefaultFormDataType.GetDefaultFormData(form.Data)
		} else if getDefaultFormDataType, ok := formService.(domain.GetDefaultFormDataWithContext); ok {
			form.Data = getDefaultFormDataType.GetDefaultFormDataWithContext(ctx, form.Data)
		}
	}
	return form, nil
}

//GetUnsubmittedForm: Use this if you need an unsubmitted form
func GetUnsubmittedForm(ctx context.Context, r *web.Request, service domain.FormService) (domain.Form, error) {
	form := domain.Form{}

	if defaultFormDataService, ok := service.(domain.GetDefaultFormData); ok {
		form.Data = defaultFormDataService.GetDefaultFormData(form.Data)
	} else if defaultFormDataService, ok := service.(domain.GetDefaultFormDataWithContext); ok {
		form.Data = defaultFormDataService.GetDefaultFormDataWithContext(ctx, form.Data)
	}

	form.ValidationRules = extractValidationRules(form)

	return form, nil
}

/*
SimpleProcessFormRequest: Parses Post Values and returns a Form object with ALL the submitted data as simple map (string of strings)
can be used if you dont need or want advanced form processing and validation.
This method don't need a "domain.FormService"
*/
func SimpleProcessFormRequest(ctx context.Context, r *web.Request) (domain.Form, error) {
	var err error
	var urlValues url.Values
	form := domain.Form{}

	if r.Request().Method != "POST" {
		form.IsSubmitted = false
		form.ValidationInfo.IsValid = true
		return form, nil
	}

	form.IsSubmitted = true

	urlValues, err = getPostValues(ctx, r)
	if err != nil {
		form.ValidationInfo.AddGeneralUnknownError(err)
		return form, err
	}
	form.ValidationInfo.IsValid = true
	dataMap := make(map[string]string)
	for k, v := range urlValues {
		dataMap[k] = strings.Join(v, " ")
	}
	form.ValidationInfo.IsValid = true
	form.Data = dataMap

	return form, nil
}

/*
ValidationErrorsToValidationInfo

Use this if you want to convert a error object to the domain.ValidationInfo

Its main purpose is to be used with the package @see gopkg.in/go-playground/validator.v9 (InvalidValidationError / ValidationErrors )

*/
func ValidationErrorsToValidationInfo(err error) domain.ValidationInfo {
	var validationInfo domain.ValidationInfo

	validationInfo.IsValid = true
	validationInfo.FieldErrors = make(map[string][]domain.Error)

	if err == nil {
		return validationInfo
	}

	if err1, ok := err.(*validator.InvalidValidationError); ok {
		validationInfo.IsValid = false
		validationInfo.AddGeneralUnknownError(err1)
	}
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErrors {
			err.Tag()
			var errorValue domain.Error
			validationInfo.IsValid = false
			fieldName := getRelativeFieldNameFromValidationError(err)
			errorValue = domain.Error{
				Tag:          err.Tag(),
				MessageKey:   "formerror_" + fieldName + "_" + err.Tag(),
				DefaultLabel: err.Field() + " wrong",
			}
			validationInfo.FieldErrors[fieldName] = append(validationInfo.FieldErrors[fieldName], errorValue)
		}
	}

	return validationInfo
}

func getRelativeFieldNameFromValidationError(err validator.FieldError) string {
	var result []string
	fieldName := err.Namespace()
	//first part of namespace is not required to have the relative path:
	fieldName = fieldName[(strings.Index(fieldName, ".") + 1):]
	for _, part := range strings.Split(fieldName, ".") {
		result = append(result, strings.ToLower(part[0:1])+part[1:])
	}
	return strings.Join(result, ".")
}

func getPostValues(ctx context.Context, r *web.Request) (url.Values, error) {
	err := r.Request().ParseForm()
	if err != nil {
		return r.Request().Form, err
	}
	return r.Request().Form, nil
}

func parseFormData(ctx context.Context, r *web.Request, values url.Values, service domain.FormService) (interface{}, error) {
	formData, err := service.ParseFormData(ctx, r, values)
	if err != nil {
		return formData, err
	}
	return formData, nil
}

func extractValidationRules(form domain.Form) map[string][]domain.ValidationRule {
	result := map[string][]domain.ValidationRule{}

	if form.Data == nil {
		return result
	}

	typeOf := reflect.TypeOf(form.Data)

	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)

		validation := field.Tag.Get("validate")
		if validation == "" {
			continue
		}

		name := field.Tag.Get("form")
		if name == "-" {
			continue
		}

		rules := strings.Split(validation, ",")
		for _, rule := range rules {
			values := strings.Split(rule, "=")
			if len(values) == 0 {
				continue
			}
			if values[0] == "omitempty" {
				continue
			}

			validationRule := domain.ValidationRule{
				Name: values[0],
			}
			if len(values)  > 1 {
				validationRule.Value = values[1]
			}

			result[name] = append(result[name], validationRule)
		}
	}

	return result
}
