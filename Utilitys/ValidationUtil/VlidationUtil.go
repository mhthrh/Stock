package ValidationUtil

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/pborman/uuid"
	"regexp"
)

type ValidationError struct {
	validator.FieldError
}

func (v ValidationError) Error() string {
	return fmt.Sprintf(
		"Key: '%s' Error: Field validation for '%s' failed on the '%s' tag",
		v.Namespace(),
		v.Field(),
		v.Tag(),
	)
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Errors() []string {
	errs := []string{}
	for _, err := range v {
		errs = append(errs, err.Error())
	}

	return errs
}

type Validation struct {
	validate *validator.Validate
}

func NewValidation() *Validation {
	validate := validator.New()
	validate.RegisterValidation("sku", validateSKU)
	validate.RegisterValidation("uuid", ValidUUID)
	return &Validation{validate}
}

func (v *Validation) Validate(i interface{}) ValidationErrors {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	var returnErrs []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		ve := ValidationError{err.(validator.FieldError)}
		returnErrs = append(returnErrs, ve)
	}

	return returnErrs
}

func validateSKU(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`[a-z]+-[a-z]+-[a-z]+`)
	sku := re.FindAllString(fl.Field().String(), -1)

	if len(sku) == 1 {
		return true
	}
	return false
}
func ValidUUID(fl validator.FieldLevel) bool {
	uuid := uuid.Parse(fl.Field().String())
	if uuid == nil {
		return false
	}
	return true
}
