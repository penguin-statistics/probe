package server

import (
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Validator is an instance used to provide interface function (conforming to echo.Validator) to validate data
type Validator struct {
}

// Validate validates data from i and return error accordingly if such data fails the validators defined in its struct
func (cv *Validator) Validate(i interface{}) error {
	valid, err := govalidator.ValidateStruct(i)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if !valid {
		return echo.NewHTTPError(http.StatusBadRequest, "unknown validator error triggered")
	}
	return nil
}
