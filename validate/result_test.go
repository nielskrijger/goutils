package validate_test

import (
	"testing"

	"github.com/nielskrijger/goutils/validate"
	"github.com/stretchr/testify/assert"
)

func TestValidationResult_Empty(t *testing.T) {
	res := validate.NewResult(nil, nil)
	res.AddError(nil)
	res.AddErrors(validate.FieldErrors{})

	assert.True(t, res.IsValid())
	assert.Len(t, res.Errors, 0)
	assert.Equal(t, "", res.Error())
	assert.True(t, res.Err() == nil)
}

func TestValidationResult_Invalid(t *testing.T) {
	err1 := validate.FieldError{"error 1", "description 1"}
	err2 := validate.FieldError{"error 2", "description 2"}

	res := validate.NewResult(err1, err2)

	assert.False(t, res.IsValid())
	assert.Len(t, res.Errors, 2)
	assert.Equal(t, res.Errors[0], err1)
	assert.Equal(t, res.Errors[1], err2)
}

func TestValidationResult_AddErrors(t *testing.T) {
	err1 := validate.FieldError{"error 1", "description 1"}
	err2 := validate.FieldError{"error 2", "description 2"}
	err3 := validate.FieldErrors{err1, err2}
	res := validate.NewResult()

	res.AddErrors(err1, err2, err3)

	assert.False(t, res.IsValid())
	assert.Len(t, res.Errors, 4)
	assert.Equal(t, err1, res.Errors[0])
	assert.Equal(t, err2, res.Errors[1])
	assert.Equal(t, err1, res.Errors[2])
	assert.Equal(t, err2, res.Errors[3])
}
