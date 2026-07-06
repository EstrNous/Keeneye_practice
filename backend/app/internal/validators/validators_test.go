package validators

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCustomValidators(t *testing.T) {
	v := validator.New()
	assert.NoError(t, RegisterCustomValidations(v))

	type sample struct {
		Email string `validate:"customemail"`
		Phone string `validate:"customphone"`
	}

	ok := sample{Email: "user@example.com", Phone: "+79001234567"}
	assert.NoError(t, v.Struct(ok))

	bad := sample{Email: "bad", Phone: "123"}
	assert.Error(t, v.Struct(bad))
}
