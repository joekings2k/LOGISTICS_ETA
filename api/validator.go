package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/joekings2k/logistics-eta/util"
)
var ValidRoles  validator.Func = func(fl validator.FieldLevel) bool {
	if role, ok := fl.Field().Interface().(string);ok{
		return util.Role(role).IsValid()
	}
	return false
}