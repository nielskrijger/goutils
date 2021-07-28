package validate_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nielskrijger/goutils/validate"
	"github.com/stretchr/testify/assert"
)

type requiredStruct struct {
	String    string            `validate:"required"`
	Array     [4]string         `validate:"required"`
	Slice     []string          `validate:"required"`
	Map       map[string]string `validate:"required"`
	Int       int               `validate:"required"`
	Int8      int8              `validate:"required"`
	Int16     int16             `validate:"required"`
	Int32     int32             `validate:"required"`
	Int64     int64             `validate:"required"`
	Uint      uint              `validate:"required"`
	Uint8     uint8             `validate:"required"`
	Uint16    uint16            `validate:"required"`
	Uint32    uint32            `validate:"required"`
	Uint64    uint64            `validate:"required"`
	Float32   float32           `validate:"required"`
	Float64   float64           `validate:"required"`
	Interface interface{}       `validate:"required"`
	Pointer   *struct{}         `validate:"required"`
	Bool      bool              `validate:"required"`
	Chan      chan string       `validate:"required"`
	Struct    struct{}          `validate:"required"`
}

func TestStruct_Required(t *testing.T) {
	tests := []string{
		"String",
		"Slice", "Map",
		"Int", "Int8", "Int16", "Int32", "Int64",
		"Uint", "Uint8", "Uint16", "Uint32", "Uint64",
		"Float32", "Float64",
		"Interface", "Pointer",
		"Bool",
		"Chan",
	}

	errs := validate.Struct(&requiredStruct{})

	var fieldErrors validate.FieldErrors

	assert.ErrorAs(t, errs, &fieldErrors)
	assert.Len(t, errs, len(tests))

	for _, err := range fieldErrors {
		assert.Contains(t, tests, err.Field)
		assert.Equal(t, err.Field+" is required", err.Description)
	}
}

type simpleStruct struct {
	A int `validate:"required"`
}

func TestStruct_SingleError(t *testing.T) {
	v := validate.NewValidator(validate.WithStandardRules())

	errs := v.Struct(&simpleStruct{})

	assert.Len(t, errs, 1)
	assert.Equal(t, "field is invalid: A", errs.Error())
}

type complexStruct struct {
	A   int `validate:"required"`
	Sub struct {
		A    int `validate:"required"`
		B    string
		C    float64 `validate:"required"`
		D    *string `validate:"required"`
		Sub2 struct {
			A int `validate:"required"`
		}
	}
}

func TestStruct_MultipleErrors(t *testing.T) {
	v := validate.NewValidator(validate.WithFullErrorPath(), validate.WithStandardRules())

	errs := v.Struct(&complexStruct{})

	assert.Len(t, errs, 5)
	assert.Equal(t, "fields are invalid: A, Sub.A, Sub.C, Sub.D, Sub.Sub2.A", errs.Error())
}

func TestStruct_WithoutFullErrorPath(t *testing.T) {
	errs := validate.Struct(&complexStruct{})

	assert.Len(t, errs, 5)
	assert.Equal(t, "fields are invalid: A, A, C, D, A", errs.Error())
}

func TestField_Required(t *testing.T) {
	err := validate.Field("", "Name", "required")

	var fieldError validate.FieldError

	assert.ErrorAs(t, err, &fieldError)
	assert.Equal(t, "field is invalid: Name", fieldError.Error())
	assert.Equal(t, "Name", fieldError.Field)
	assert.Equal(t, "Name is required", fieldError.Description)
}

func TestField_Optional(t *testing.T) {
	err := validate.Field("", "Name", "optional,lte=3")

	assert.Nil(t, err)
}

func (u *fakeUser) Validate() error {
	return validate.Fields( // nolint:wrapcheck
		validate.Field(u.Name, "Name", "required,gte=3,lte=25"),
		validate.Field(u.Gender, "Gender", "gender"),
	)
}

var fakeValidUser = &fakeUser{
	Name:   "John Doe",
	Gender: "male",
}

func TestFields_Valid(t *testing.T) {
	assert.True(t, fakeValidUser.Validate() == nil)
}

var fakeInvalidUser = &fakeUser{
	Name:   "Jo",
	Gender: "invalid",
}

func TestFields_Invalid(t *testing.T) {
	errs := fakeInvalidUser.Validate()

	var fieldErrors validate.FieldErrors

	assert.ErrorAs(t, errs, &fieldErrors)
	assert.Len(t, fieldErrors, 2)
}

func TestGTE(t *testing.T) {
	tests := []struct {
		test  interface{}
		error string
	}{
		{"12", "Value must be at least 3 characters long"},
		{"123", ""},
		{2, "Value must be at least 3"},
		{3, ""},
		{uint(2), "Value must be at least 3"},
		{uint(3), ""},
		{2.999999, "Value must be at least 3"},
		{3.0, ""},
		{[]string{"a", "b"}, "Value must contain at least 3 elements"},
		{[]string{"a", "b", "c"}, ""},
	}

	for _, tt := range tests {
		err := validate.Field(tt.test, "Value", "gte=3")
		if tt.error == "" {
			assert.Nil(t, err, fmt.Sprintf("failed validation for %+v", tt.test))
		} else {
			assert.NotNil(t, err)

			var fieldError validate.FieldError
			assert.ErrorAs(t, err, &fieldError)
			assert.Equal(t, tt.error, fieldError.Description)
		}
	}
}

func TestLTE(t *testing.T) {
	tests := []struct {
		test  interface{}
		error string
	}{
		{"12", ""},
		{"123", "Value must be at most 2 characters long"},
		{2, ""},
		{3, "Value maximum value is 2"},
		{uint(2), ""},
		{uint(3), "Value maximum value is 2"},
		{2.0, ""},
		{2.000001, "Value maximum value is 2"},
		{[]string{"a", "b"}, ""},
		{[]string{"a", "b", "c"}, "Value may not contain more than 2 elements"},
	}

	for _, tt := range tests {
		err := validate.Field(tt.test, "Value", "lte=2")
		if tt.error == "" {
			assert.Nil(t, err, fmt.Sprintf("failed validation for %+v", tt.test))
		} else {
			assert.NotNil(t, err)

			var fieldError validate.FieldError
			assert.ErrorAs(t, err, &fieldError)
			assert.Equal(t, tt.error, fieldError.Description)
		}
	}
}

var invalidTypeTests = []string{"gte", "lte"}

type testStruct struct{}

func TestGLTE_InvalidType(t *testing.T) {
	for _, tag := range invalidTypeTests {
		assert.PanicsWithValue(t, "invalid type for "+tag+" tag", func() {
			_ = validate.Field(&testStruct{}, "TEST", tag+"=3")
		})
	}
}

type fakeUser struct {
	Birthdate       *time.Time `validate:"isodate,mindate=1900-01-01,maxdate=2010-12-31"`
	Subject         string     `validate:"resourcename"`
	Subjects        []string   `validate:"resourcename"`
	Resource        string     `validate:"resourcepattern"`
	Resources       []string   `validate:"resourcepattern"`
	Gender          string     `validate:"gender"`
	Future          *time.Time `validate:"mindate=now"`
	Past            *time.Time `validate:"maxdate=now"`
	Az              string     `validate:"az_"`
	Azs             []string   `validate:"az_"`
	AZ09            string     `validate:"aZ09_"`
	AZ09s           []string   `validate:"aZ09_"`
	Name            string     `validate:"name"`
	Names           []string   `validate:"name"`
	Zoneinfo        string     `validate:"zoneinfo"`
	Locale          string     `validate:"locale"`
	BirthdateString string     `validate:"isodate,mindate=1900-01-01,maxdate=2010-12-31"`
	URL             string     `validate:"url"`
	PrimaryEmail    *fakeEmail `validate:"optional"`
}

type fakeEmail struct {
	Email string `validate:"email"`
}

var (
	dateZero     = time.Unix(0, 0)
	dateWithTime = time.Date(2001, 1, 2, 3, 5, 6, 7, time.UTC)
	dateLongAgo  = time.Date(1899, 11, 17, 0, 0, 0, 0, time.UTC)
	date2011     = time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC)
	yesterday    = time.Now().UTC().Add(-24 * time.Hour)
	tomorrow     = time.Now().UTC().Add(24 * time.Hour)
	now          = time.Now().UTC()
)

var ruleTests = []struct {
	user   interface{}
	errors map[string]string
}{
	// date
	{&fakeUser{Birthdate: nil}, nil},
	{&fakeUser{Birthdate: &dateZero}, map[string]string{
		"Birthdate": "Birthdate is not a valid date (YYYY-MM-DD)",
	}},
	{&fakeUser{Birthdate: &dateWithTime}, map[string]string{
		"Birthdate": "Birthdate is not a valid date (YYYY-MM-DD)",
	}},
	{&fakeUser{BirthdateString: ""}, nil},
	{&fakeUser{BirthdateString: "2010-01-02"}, nil},
	{&fakeUser{BirthdateString: "01-02-2006"}, map[string]string{
		"BirthdateString": "BirthdateString is not a valid date (YYYY-MM-DD)",
	}},
	{&fakeUser{BirthdateString: "2010-01-02T23:33Z"}, map[string]string{
		"BirthdateString": "BirthdateString is not a valid date (YYYY-MM-DD)",
	}},

	// mindate
	{&fakeUser{Birthdate: &dateLongAgo}, map[string]string{
		"Birthdate": "Birthdate minimum date is 1900-01-01",
	}},
	{&fakeUser{Future: &yesterday}, map[string]string{
		"Future": "Future minimum date is " + now.Format("2006-01-02"),
	}},
	{&fakeUser{BirthdateString: "1899-12-31"}, map[string]string{
		"BirthdateString": "BirthdateString minimum date is 1900-01-01",
	}},

	// maxdate
	{&fakeUser{Birthdate: &date2011}, map[string]string{
		"Birthdate": "Birthdate maximum date is 2010-12-31",
	}},

	// gender
	{&fakeUser{Gender: ""}, nil},
	{&fakeUser{Gender: "male"}, nil},
	{&fakeUser{Gender: "m"}, map[string]string{
		"Gender": "Gender must be either male, female, genderqueer",
	}},
	{&fakeUser{Past: &tomorrow}, map[string]string{
		"Past": "Past maximum date is " + now.Format("2006-01-02"),
	}},

	// az_
	{&fakeUser{Az: ""}, nil},
	{&fakeUser{Az: "t_est_"}, nil},
	{&fakeUser{Az: "_test"}, map[string]string{
		"Az": "Az must contain a-z, _ and not start with a _",
	}},
	{&fakeUser{Az: "te0st"}, map[string]string{
		"Az": "Az must contain a-z, _ and not start with a _",
	}},
	{&fakeUser{Azs: []string{"test", "test__"}}, nil},
	{&fakeUser{Azs: []string{"Test", "test0"}}, map[string]string{
		"Azs": "Azs must contain a-z, _ and not start with a _",
	}},

	// aZ09_
	{&fakeUser{AZ09: ""}, nil},
	{&fakeUser{AZ09: "Test09__"}, nil},
	{&fakeUser{AZ09: "_test"}, map[string]string{
		"AZ09": "AZ09 must contain 0-9, A-Z, _ and not start with a _",
	}},
	{&fakeUser{AZ09: "te st"}, map[string]string{
		"AZ09": "AZ09 must contain 0-9, A-Z, _ and not start with a _",
	}},
	{&fakeUser{AZ09s: []string{"0_9aZ", "Test09__"}}, nil},
	{&fakeUser{AZ09s: []string{"0_9aZ", "Test09__", "_test"}}, map[string]string{
		"AZ09s": "AZ09s must contain 0-9, A-Z, _ and not start with a _",
	}},

	// name
	{&fakeUser{Name: ""}, nil},
	{&fakeUser{Name: "ŵƼǗǨȐ ,.'- ȣΏШア艮"}, nil},
	{&fakeUser{Name: " spacebegin"}, map[string]string{
		"Name": "Name must contain unicode letters -,.' and not start or end with a space",
	}},
	{&fakeUser{Name: "no9"}, map[string]string{
		"Name": "Name must contain unicode letters -,.' and not start or end with a space",
	}},
	{&fakeUser{Names: []string{"Doe John", "John Doe"}}, nil},
	{&fakeUser{Names: []string{"hello", "09", "hi"}}, map[string]string{
		"Names": "Names must contain unicode letters -,.' and not start or end with a space",
	}},

	// zoneinfo
	{&fakeUser{Zoneinfo: ""}, nil},
	{&fakeUser{Zoneinfo: "Europe/Amsterdam"}, nil},
	{&fakeUser{Zoneinfo: "Unknown/Europe"}, map[string]string{
		"Zoneinfo": "Zoneinfo is not a valid zoneinfo string (example: 'Europe/Amsterdam')",
	}},

	// locale
	{&fakeUser{Locale: ""}, nil},
	{&fakeUser{Locale: "en und-u-cu-USD zh_Hant_HK_u_co_pinyi"}, nil},
	{&fakeUser{Locale: "en en-u"}, map[string]string{
		"Locale": "Locale must contain BCP47 language tags separated by spaces",
	}},

	// url
	{&fakeUser{URL: ""}, nil},
	{&fakeUser{URL: "https://www.cnn.com/test?test=bliep#hashtag=123"}, nil},
	{&fakeUser{URL: "http://foobar.com/t$-_.+!*\\'(),"}, nil},
	{&fakeUser{URL: "//www.cnn.com/test?test=bliep#hashtag=123"}, map[string]string{
		"URL": "URL is not a valid url",
	}},

	// email
	{&fakeUser{PrimaryEmail: &fakeEmail{Email: ""}}, nil},
	{&fakeUser{PrimaryEmail: &fakeEmail{Email: "Dörte@Sörensen.example.com"}}, nil},
	{&fakeUser{PrimaryEmail: &fakeEmail{Email: "not valid"}}, map[string]string{
		"Email": "Email is not a valid email",
	}},

	// resourcename
	{&fakeUser{Subject: ""}, nil},
	{&fakeUser{Subject: "mtx:account:test/1-2"}, nil},
	{&fakeUser{Subject: "mtx:accounT"}, map[string]string{
		"Subject": "Subject must start with 'mtx:' and may contain: a-z, 0-9, -, /, and :",
	}},
	{&fakeUser{Subject: "mtx:"}, map[string]string{
		"Subject": "Subject must start with 'mtx:' and may contain: a-z, 0-9, -, /, and :",
	}},
	{&fakeUser{Subject: "mtx:test*"}, map[string]string{
		"Subject": "Subject must start with 'mtx:' and may contain: a-z, 0-9, -, /, and :",
	}},
	{&fakeUser{Subjects: []string{"", "mtx:account:test-1234", "mtx:test:0-9"}}, nil},
	{&fakeUser{Subjects: []string{"mtx:test*", "mtx:account:test-1234", "mtx:test:0-9"}}, map[string]string{
		"Subjects": "Subjects must start with 'mtx:' and may contain: a-z, 0-9, -, /, and :",
	}},

	// resourcenamepattern
	{&fakeUser{Resource: ""}, nil},
	{&fakeUser{Resource: "mtx:*:b12*:c/*d:**"}, nil},
	{&fakeUser{Resource: "mtx:accounT"}, map[string]string{
		"Resource": "Resource must start with 'mtx:' and may contain: a-z, 0-9, -, /, *, and :",
	}},
	{&fakeUser{Resources: []string{"", "mtx:account:*:123"}}, nil},
	{&fakeUser{Resources: []string{"", "mtx:account:*:123", "mtx:no_underscore"}}, map[string]string{
		"Resources": "Resources must start with 'mtx:' and may contain: a-z, 0-9, -, /, *, and :",
	}},
}

func TestRules(t *testing.T) {
	for _, tt := range ruleTests {
		errs := validate.Struct(tt.user)
		if tt.errors == nil {
			assert.Nil(t, errs, fmt.Sprintf("failed validation for %+v", tt.user))
		} else {
			for field, desc := range tt.errors {
				var fieldErrors validate.FieldErrors
				assert.ErrorAs(t, errs, &fieldErrors)
				err := findError(fieldErrors, field)

				assert.NotNil(t, err, "failed "+field+" validation")
				assert.Equal(t, desc, err.Description)
			}
		}
	}
}

func findError(errs validate.FieldErrors, field string) validate.FieldError {
	for _, err := range errs {
		if err.Field == field {
			return err
		}
	}

	panic(fmt.Sprintf("field %q not found", field))
}

func TestAliases(t *testing.T) {
	tests := []struct {
		value interface{}
		tag   string
		field string
		error string
	}{
		// username
		{"Hello123_", "username", "Username", ""},
		{"te", "username", "Username", "Username must be at least 4 characters long"},
		{strings.Repeat("a", 21), "username", "Username", "Username must be at most 20 characters long"},
		{"_rudolph", "username", "Username", "Username must contain 0-9, A-Z, _ and not start with a _"},

		// birthdate
		{"", "birthdate", "Birthdate", ""},
		{"2010-01-02", "birthdate", "Birthdate", ""},
		{tomorrow.Format("2006-01-02"), "birthdate", "Birthdate", "Birthdate maximum date is " + now.Format("2006-01-02")},
		{"1899-12-31", "birthdate", "Birthdate", "Birthdate minimum date is 1900-01-01"},
		{"1899-12-31T12:30", "birthdate", "Birthdate", "Birthdate is not a valid date (YYYY-MM-DD)"},
		{validate.InvalidTime, "birthdate", "Birthdate", "Birthdate is not a valid date (YYYY-MM-DD)"},
	}

	for _, tt := range tests {
		err := validate.Field(tt.value, tt.field, tt.tag)
		if tt.error == "" {
			assert.Nil(t, err)
		} else {
			var fieldError validate.FieldError
			assert.ErrorAs(t, err, &fieldError)
			assert.Equal(t, tt.error, fieldError.Description)
		}
	}
}

func TestRules_InvalidTypes(t *testing.T) {
	tests := []struct {
		value interface{}
		tags  string
		error string
	}{
		{false, "isodate", "invalid type for isodate tag"},
		{true, "mindate=2011-01-02", "invalid type for mindate tag"},
		{false, "maxdate=2011-01-02", "invalid type for maxdate tag"},
		{false, "name", "invalid type for name tag"},
		{false, "", "unknown validate tag \"\""},
		{false, "unknown", "unknown validate tag \"unknown\""},
		{false, "zoneinfo", "invalid type for zoneinfo tag"},
		{false, "locale", "invalid type for locale tag"},
		{false, "aZ09_", "invalid type for aZ09_ tag"},
		{false, "gender", "invalid type for gender tag"},
	}

	for _, tt := range tests {
		assert.PanicsWithValue(t, tt.error, func() {
			_ = validate.Field(tt.value, "TEST", tt.tags)
		})
	}
}
