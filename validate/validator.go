// Package validator implements value validations
//
// Copyright 2014 Roberto Teixeira <robteix@robteix.com>
// Copyright 2019 Niels Krijger <niels@kryger.nl>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validate

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

var (
	// ErrUnsupported is the error error returned when a validation rule
	// is used with an unsupported variable type.
	ErrUnsupported = errors.New("unsupported type")

	tagCache   = sync.Map{}
	sepPattern = regexp.MustCompile(`((?:^|[^\\])(?:\\\\)*),`)
)

// FieldErrors contains an array of errors returned by the validation
// functions.
type FieldErrors []FieldError

// FieldErrors implements the Error interface and returns the first error as
// string if existent.
func (ve FieldErrors) Error() string {
	err := ""
	for _, fe := range ve {
		if err != "" {
			err += ", "
		}

		err += fe.Field
	}

	if len(ve) == 1 {
		return "field is invalid: " + err
	}

	return "fields are invalid: " + err
}

// FieldError contains an error message for a given field.
type FieldError struct {
	Field       string
	Description string
}

// Error implements the Error interface.
func (fe FieldError) Error() string {
	return "field is invalid: " + fe.Field
}

// ValidationRule specifies the validation functions ("Checkers")
// and error message function ("ErrorFunc") for a given Tag.
//
// The Checker and ErrorFunc are defined separately for more flexibility.
type ValidationRule struct {
	// Tag is the struct tag to run this rule on.
	Tag string

	// Checker returns true when a value is valid, otherwise false.
	Checker RuleChecker

	// ErrorFunc is called when Checker returned false. The
	// ErrorFunc returns a proper error message.
	ErrorFunc RuleErrorFunc
}

// RuleChecker is a function that receives the value of a
// field and a parameter used for the respective validation tag.
// Returns true when validation passed, or false if it didn't.
type RuleChecker func(v interface{}, param string) bool

// RuleErrorFunc returns an error message. This function is
// called when RuleChecker returned false.
type RuleErrorFunc func(field string, value interface{}, tag Tag) string

// Validator is the main validation construct.
type Validator struct {
	tagName       string
	rules         map[string]ValidationRule
	fullErrorPath bool
	tagAliases    map[string][]Tag
}

var DefaultValidator = NewValidator(
	WithStandardRules(),
	WithStandardAliases())

// Option allows functional options.
type Option func(*Validator)

// WithFullErrorPath adds the entire path in a struct or map
// to the field error.
func WithFullErrorPath() func(*Validator) {
	return func(v *Validator) {
		v.fullErrorPath = true
	}
}

// WithStandardRules adds the packaged validation rules.
func WithStandardRules() func(*Validator) {
	return func(v *Validator) {
		for _, rule := range StandardRules {
			v.AddRule(rule)
		}
	}
}

// WithStandardRules adds the packaged tag aliases.
func WithStandardAliases() func(*Validator) {
	return func(v *Validator) {
		for alias, tags := range StandardAliases {
			v.AddAlias(alias, tags)
		}
	}
}

// NewValidator creates a new Validator with all default
// validation rules.
func NewValidator(options ...Option) *Validator {
	val := &Validator{
		tagName:    "validate",
		rules:      map[string]ValidationRule{},
		tagAliases: make(map[string][]Tag),
	}
	for _, option := range options {
		option(val)
	}

	return val
}

// AddRule adds a new rule or overwrites and existing rule
// if a rule with the same tag already exists.
func (mv *Validator) AddRule(rule ValidationRule) {
	mv.rules[rule.Tag] = rule
}

// AddAlias adds a new alias or overwrites an existing one
// if alias already exists. Panics if one of the tags
// does not exist.
func (mv *Validator) AddAlias(alias string, tags string) {
	mv.tagAliases[alias] = mv.mustParseTags(tags)
}

// Struct validates the fields of a struct based on
// the validator's tag and returns an array FieldErrors if
// one or more errors were found. Panics if value is not
// a struct.
func Struct(value interface{}) error {
	return DefaultValidator.Struct(value)
}

// Struct validates the fields of a struct based on
// the validator's tag and returns an array FieldErrors if
// one or more errors were found. Returns nil if no errors
// were found.
//
// Panics if given value is not a struct.
func (mv *Validator) Struct(value interface{}) error {
	errs := mv.validateStruct(value, "")
	if len(errs) > 0 {
		return errs
	}

	return nil
}

// Struct validates the fields of a struct based on
// the validator's tag and returns an array FieldErrors if
// one or more errors were found. Returns nil if no errors
// were found.
//
// Panics if given value is not a struct.
func (mv *Validator) validateStruct(value interface{}, fieldName string) (errs FieldErrors) {
	sv := reflect.ValueOf(value)
	st := reflect.TypeOf(value)

	if sv.Kind() == reflect.Ptr {
		if sv.IsNil() {
			return nil
		}

		errs = mv.validateStruct(sv.Elem().Interface(), fieldName)
	} else {
		errs = mv.validateStructFields(st, sv)
	}

	if len(errs) == 0 {
		return nil
	}

	if !mv.fullErrorPath || fieldName == "" {
		return errs
	}

	result := make(FieldErrors, 0, len(errs))

	// Prefix field name to returned error details, e.g. "user.firstname" instead of just "firstname"
	for _, err := range errs {
		result = append(result, FieldError{
			Field:       fieldName + "." + err.Field,
			Description: err.Description,
		})
	}

	return result
}

func (mv *Validator) validateStructFields(st reflect.Type, sv reflect.Value) (result FieldErrors) {
	fieldCount := sv.NumField()
	for i := 0; i < fieldCount; i++ {
		field := st.Field(i).Name

		// only public fields are validatable
		if !unicode.IsUpper(rune(field[0])) {
			continue
		}

		f := sv.Field(i)

		// deal with pointers
		for f.Kind() == reflect.Ptr && !f.IsNil() {
			f = f.Elem()
		}

		// fetch tag to validate
		tag := st.Field(i).Tag.Get(mv.tagName)
		if tag == "-" {
			continue
		}

		if tag != "" {
			// tags are only defined on validatable fields
			if err := mv.Field(f.Interface(), st.Field(i).Name, tag); err != nil {
				var fieldError FieldError

				errors.As(err, &fieldError)
				result = append(result, fieldError)
			}
		}

		// validate struct, interface, array, slice or map that have no tag
		errs := mv.deepValidateTaglessField(f, field)
		if errs != nil {
			result = append(result, errs...)
		}
	}

	return result
}

// deepValidateTaglessField validates a struct, interface, array, slice or map that have no tag.
func (mv *Validator) deepValidateTaglessField(value reflect.Value, field string) FieldErrors {
	switch value.Kind() {
	case reflect.Interface, reflect.Ptr:
		if value.IsNil() {
			// Whenever nil value is passed there is nothing to validate further
			return nil
		}

		fallthrough
	case reflect.Struct:
		return mv.validateStruct(value.Interface(), field)
	case reflect.Array, reflect.Slice:
		return mv.validateCollection(value, field)
	case reflect.Map:
		return mv.validateMap(value, field)
	default:
	}

	return nil
}

func (mv *Validator) validateCollection(value reflect.Value, field string) (result FieldErrors) {
	for i := 0; i < value.Len(); i++ {
		if errs := mv.deepValidateTaglessField(value.Index(i), field+"["+string(rune(i))+"]"); errs != nil {
			if result == nil {
				result = FieldErrors{}
			}

			result = append(result, errs...)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func (mv *Validator) validateMap(value reflect.Value, field string) (result FieldErrors) {
	for _, key := range value.MapKeys() {
		// validate the map key
		errs := mv.deepValidateTaglessField(key, fmt.Sprintf("%s[%+v](key)", field, key.Interface()))
		if errs != nil {
			if result == nil {
				result = FieldErrors{}
			}

			result = append(result, errs...)
		}

		// validate the map value
		value := value.MapIndex(key)

		errs = mv.deepValidateTaglessField(value, fmt.Sprintf("%s[%+v](value)", field, key.Interface()))
		if errs != nil {
			if result == nil {
				result = FieldErrors{}
			}

			result = append(result, errs...)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// Field validates a value based on the provided tags. Returns the
// first error found or nil when valid.
func Field(val interface{}, field string, tags string) error {
	return DefaultValidator.Field(val, field, tags)
}

// Field validates a value based on the provided tags. Returns the
// first error found or nil when valid.
func (mv *Validator) Field(val interface{}, field string, tags string) error {
	if tags == "-" {
		return nil
	}

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return mv.Field(v.Elem().Interface(), field, tags)
	}

	var err error

	switch v.Kind() {
	case reflect.Invalid:
		err = mv.singleField(nil, field, tags)
	default:
		err = mv.singleField(val, field, tags)
	}

	return err
}

// singleField validates one single variable.
func (mv *Validator) singleField(v interface{}, field string, tag string) error {
	tags := mv.mustParseTags(tag)
	for _, t := range tags {
		if !t.Rule.Checker(v, t.Param) {
			// The "optional" tag does not define an error function, it simply stops
			// further validation.
			if t.Rule.ErrorFunc == nil {
				return nil
			}

			return FieldError{
				Field:       field,
				Description: t.Rule.ErrorFunc(field, v, t),
			}
		}
	}

	return nil
}

// Fields is a helper method to wrap a set of validate.Field() and returns
// a FieldErrors struct.
//
// If you intend to add more errors you should consider using validate.NewResult()
// instead.
func Fields(errs ...error) error {
	var result FieldErrors

	for _, err := range errs {
		if err != nil {
			if result == nil {
				result = make([]FieldError, 0)
			}

			var fieldError FieldError

			errors.As(err, &fieldError)
			result = append(result, fieldError)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

type Tag struct {
	Name  string
	Rule  ValidationRule
	Param string
}

// mustParseTags parses all individual tags found within a tag value.
// Caches the result. Panics if an unknown tag was found.
func (mv *Validator) mustParseTags(t string) []Tag {
	if val, ok := tagCache.Load(t); ok {
		return val.([]Tag)
	}

	tl := splitUnescapedComma(t)
	tags := make([]Tag, 0, len(tl))

	for _, i := range tl {
		i = strings.ReplaceAll(i, `\,`, ",")
		tg := Tag{}
		v := strings.SplitN(i, "=", 2) //nolint:gomnd
		tg.Name = strings.Trim(v[0], " ")

		if tg.Name == "" {
			panic(fmt.Sprintf("unknown %s tag %q", mv.tagName, tg.Name))
		}

		if len(v) > 1 {
			tg.Param = strings.Trim(v[1], " ")
		}

		var found bool
		if tg.Rule, found = mv.rules[tg.Name]; !found {
			if val, ok := mv.tagAliases[tg.Name]; ok {
				tags = append(tags, val...)
			} else {
				panic(fmt.Sprintf("unknown %s tag %q", mv.tagName, tg.Name))
			}
		} else {
			tags = append(tags, tg)
		}
	}

	tagCache.Store(t, tags)

	return tags
}

func splitUnescapedComma(str string) []string {
	indexes := sepPattern.FindAllStringIndex(str, -1)
	pieces := make([]string, 0)
	last := 0

	for _, is := range indexes {
		pieces = append(pieces, str[last:is[1]-1])
		last = is[1]
	}

	pieces = append(pieces, str[last:])

	return pieces
}
