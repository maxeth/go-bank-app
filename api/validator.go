package api

import (
	"github.com/go-playground/validator/v10"
	library "github.com/maxeth/go-bank-app/library"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	val, ok := fl.Field().Interface().(string)

	if ok {
		// field is a string, check whether it is a valid currency
		return library.IsSupportedCurrency(val)
	}

	// field is not a string so it cannot be a supported currency
	return false
}
