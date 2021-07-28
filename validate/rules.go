package validate

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
)

var (
	// InvalidTime can be set on a time.Time field to indicate parsing
	// the timestamp failed due to invalid formatting.
	InvalidTime = time.Unix(0, 0)

	validGenders          = []string{"male", "female", "genderqueer"}
	regexpAz              = regexp.MustCompile(`^[a-z][a-z_]*$`)
	regexpAZ09            = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]*$`)
	regexpName            = regexp.MustCompile(`^[\p{L},.'-][\p{L} ,.'-]*[\p{L},.'-]$`)
	regexpEmail           = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:\\(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22)))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$") //nolint
	regexpResourceName    = regexp.MustCompile("^mtx:[a-z0-9-/]+(:[a-z0-9-/]+)*$")
	regexpResourcePattern = regexp.MustCompile("^mtx:[a-z0-9-*/]+(:[a-z0-9-*/]+)*$")

	StandardRules = []ValidationRule{
		{
			Tag:       "required",
			Checker:   Required,
			ErrorFunc: RequiredErr,
		},
		{
			Tag:       "optional",
			Checker:   Optional,
			ErrorFunc: nil, // No error causes validation to stop
		},
		{
			Tag:       "gte",
			Checker:   GTE,
			ErrorFunc: GTEErr,
		},
		{
			Tag:       "lte",
			Checker:   LTE,
			ErrorFunc: LTEErr,
		},
		{
			Tag:       "gender",
			Checker:   Gender,
			ErrorFunc: GenderErr,
		},
		{
			Tag:       "isodate",
			Checker:   ISODate,
			ErrorFunc: ISODateErr,
		},
		{
			Tag:       "mindate",
			Checker:   MinDate,
			ErrorFunc: MinDateErr,
		},
		{
			Tag:       "maxdate",
			Checker:   MaxDate,
			ErrorFunc: MaxDateErr,
		},
		{
			Tag:       "name",
			Checker:   Name,
			ErrorFunc: NameErr,
		},
		{
			Tag:       "az_",
			Checker:   Az,
			ErrorFunc: AzErr,
		},
		{
			Tag:       "aZ09_",
			Checker:   AZ09,
			ErrorFunc: AZ09Err,
		},
		{
			Tag:       "zoneinfo",
			Checker:   Zoneinfo,
			ErrorFunc: ZoneinfoErr,
		},
		{
			Tag:       "locale",
			Checker:   Locale,
			ErrorFunc: LocaleErr,
		},
		{
			Tag:       "url",
			Checker:   URL,
			ErrorFunc: URLErr,
		},
		{
			Tag:       "email",
			Checker:   Email,
			ErrorFunc: EmailErr,
		},
		{
			Tag:       "resourcename",
			Checker:   ResourceName,
			ErrorFunc: ResourceNameErr,
		},
		{
			Tag:       "resourcepattern",
			Checker:   ResourcePattern,
			ErrorFunc: ResourcePatternErr,
		},
	}

	StandardAliases = map[string]string{
		"username":  "aZ09_,gte=4,lte=20",
		"birthdate": "isodate,mindate=1900-01-01,maxdate=now",
	}
)

// Required tests whether a variable is non-zero as defined by
// the golang spec.
//
// You're advised not to use this validation for booleans and numbers,
// since golang defaults empty numbers to 0 and empty booleans to false.
func Required(v interface{}, _ string) bool { //nolint:cyclop
	st := reflect.ValueOf(v)
	switch st.Kind() {
	case reflect.String:
		return len(st.String()) != 0
	case reflect.Ptr, reflect.Interface:
		return !st.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return st.Len() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return st.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return st.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return st.Float() != 0
	case reflect.Bool:
		return st.Bool()
	case reflect.Invalid:
		return false // always invalid
	case reflect.Struct:
		return true // always valid since only nil pointers are empty
	default:
		return false
	}
}

func RequiredErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s is required", field)
}

// Optional tests whether a variable is zero as defined by
// the golang spec.
func Optional(v interface{}, _ string) bool {
	// It's the same as "required", causing an error when the value
	// is empty. However in this case no error will be defined and
	// the validation pipeline simply stops.
	return Required(v, "")
}

// GTE tests whether a variable value is larger or equal to a given
// number. For number types, it's a simple greater-than test; for strings
// it tests the number of characters whereas for maps and slices it tests
// the number of items.
func GTE(v interface{}, param string) bool {
	st := reflect.ValueOf(v)
	switch st.Kind() {
	case reflect.String:
		return int64(len(st.String())) >= asInt(param)
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(st.Len()) >= asInt(param)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return st.Int() >= asInt(param)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return st.Uint() >= asUint(param)
	case reflect.Float32, reflect.Float64:
		return st.Float() >= asFloat(param)
	default:
		panic("invalid type for gte tag")
	}
}

func GTEErr(field string, v interface{}, t Tag) string {
	st := reflect.ValueOf(v)
	switch st.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return fmt.Sprintf("%s must contain at least %s elements", field, t.Param)
	case reflect.String:
		return fmt.Sprintf("%s must be at least %s characters long", field, t.Param)
	default:
		return fmt.Sprintf("%s must be at least %s", field, t.Param)
	}
}

// LTE tests whether a variable value is smaller or equal to a given
// number. For number types, it's a simple lesser-than test; for strings
// it tests the number of characters whereas for maps and slices it tests
// the number of items.
func LTE(v interface{}, param string) bool {
	st := reflect.ValueOf(v)

	switch st.Kind() {
	case reflect.String:
		return int64(len(st.String())) <= asInt(param)
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(st.Len()) <= asInt(param)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return st.Int() <= asInt(param)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return st.Uint() <= asUint(param)
	case reflect.Float32, reflect.Float64:
		return st.Float() <= asFloat(param)
	default:
		panic("invalid type for lte tag")
	}
}

func LTEErr(field string, v interface{}, t Tag) string {
	st := reflect.ValueOf(v)

	switch st.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return fmt.Sprintf("%s may not contain more than %s elements", field, t.Param)
	case reflect.String:
		return fmt.Sprintf("%s must be at most %s characters long", field, t.Param)
	default:
		return fmt.Sprintf("%s maximum value is %s", field, t.Param)
	}
}

func Gender(v interface{}, _ string) bool {
	val, ok := v.(string)
	if !ok {
		panic("invalid type for gender tag")
	}

	if val == "" {
		return true
	}

	for _, gender := range validGenders {
		if gender == val {
			return true
		}
	}

	return false
}

func GenderErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must be either %s", field, strings.Join(validGenders, ", "))
}

func ISODate(v interface{}, _ string) bool {
	st := reflect.ValueOf(v)
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			return true
		}

		st = st.Elem()
	}

	switch st.Kind() {
	case reflect.String:
		if st.String() == "" {
			return true
		}

		t, err := time.Parse("2006-01-02", st.String())
		if err != nil {
			return false
		}

		res := isWholeDate(t)

		return res
	case reflect.Struct:
		if t, ok := v.(time.Time); ok {
			if t.Equal(InvalidTime) {
				return false
			}

			return isWholeDate(t)
		}

		return false
	default:
		panic("invalid type for isodate tag")
	}
}

func isWholeDate(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
}

func ISODateErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s is not a valid date (YYYY-MM-DD)", field)
}

func MinDate(v interface{}, param string) bool { //nolint:cyclop
	st := reflect.ValueOf(v)
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			return true
		}

		st = st.Elem()
	}

	switch st.Kind() {
	case reflect.String:
		if st.String() == "" {
			return true
		}

		t, err := time.Parse("2006-01-02", st.String())
		if err != nil {
			return false
		}

		minDate := parseDate(param)

		return t.After(minDate) || t.Equal(minDate)
	case reflect.Struct:
		if t, ok := v.(time.Time); ok {
			minDate := parseDate(param)

			return t.After(minDate) || t.Equal(minDate)
		}

		return false
	default:
		panic("invalid type for mindate tag")
	}
}

func MinDateErr(field string, _ interface{}, t Tag) string {
	return fmt.Sprintf("%s minimum date is %s", field, nowToDateString(t.Param))
}

func MaxDate(v interface{}, param string) bool { //nolint:cyclop
	st := reflect.ValueOf(v)
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			return true
		}

		st = st.Elem()
	}

	switch st.Kind() {
	case reflect.String:
		if st.String() == "" {
			return true
		}

		t, err := time.Parse("2006-01-02", st.String())
		if err != nil {
			return false
		}

		maxDate := parseDate(param)

		return t.Before(maxDate) || t.Equal(maxDate)
	case reflect.Struct:
		if t, ok := v.(time.Time); ok {
			maxDate := parseDate(param)

			return t.Before(maxDate) || t.Equal(maxDate)
		}

		return false
	default:
		panic("invalid type for maxdate tag")
	}
}

func MaxDateErr(field string, _ interface{}, t Tag) string {
	return fmt.Sprintf("%s maximum date is %s", field, nowToDateString(t.Param))
}

func parseDate(date string) time.Time {
	if date == "now" {
		return time.Now().UTC()
	}

	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err) // This is a coding error in the tag value
	}

	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func nowToDateString(date string) string {
	if date == "now" {
		return time.Now().UTC().Format("2006-01-02")
	}

	return date
}

func Az(v interface{}, _ string) bool {
	return RegexChecker("az_", regexpAz, v)
}

func AzErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must contain a-z, _ and not start with a _", field)
}

func AZ09(v interface{}, _ string) bool {
	return RegexChecker("aZ09_", regexpAZ09, v)
}

func AZ09Err(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must contain 0-9, A-Z, _ and not start with a _", field)
}

func Name(v interface{}, _ string) bool {
	return RegexChecker("name", regexpName, v)
}

func NameErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must contain unicode letters -,.' and not start or end with a space", field)
}

func Zoneinfo(v interface{}, _ string) bool {
	val, ok := v.(string)
	if !ok {
		panic("invalid type for zoneinfo tag")
	}

	if val != "" {
		_, err := time.LoadLocation(val)
		if err != nil {
			return false
		}
	}

	return true
}

func ZoneinfoErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s is not a valid zoneinfo string (example: 'Europe/Amsterdam')", field)
}

func Locale(v interface{}, _ string) bool {
	val, ok := v.(string)
	if !ok {
		panic("invalid type for locale tag")
	}

	if val != "" {
		tags := strings.Split(val, " ")
		for _, s := range tags {
			_, err := language.Parse(s)
			if err != nil {
				return false
			}
		}
	}

	return true
}

func LocaleErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must contain BCP47 language tags separated by spaces", field)
}

func URL(v interface{}, _ string) bool {
	val, ok := v.(string)
	if !ok {
		panic("invalid type for url tag")
	}

	if val == "" {
		return true
	}

	// Strip # prior to validation
	var i int
	if i = strings.Index(val, "#"); i > -1 {
		val = val[:i]
	}

	parsedURL, err := url.ParseRequestURI(val)
	if err != nil || parsedURL.Scheme == "" {
		return false
	}

	return true
}

func URLErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s is not a valid url", field)
}

func Email(v interface{}, _ string) bool {
	return RegexChecker("email", regexpEmail, v)
}

func EmailErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s is not a valid email", field)
}

func ResourceName(v interface{}, _ string) bool {
	return RegexChecker("resourcename", regexpResourceName, v)
}

func ResourceNameErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must start with 'mtx:' and may contain: a-z, 0-9, -, /, and :", field)
}

func ResourcePattern(v interface{}, _ string) bool {
	return RegexChecker("resourcepattern", regexpResourcePattern, v)
}

func ResourcePatternErr(field string, _ interface{}, _ Tag) string {
	return fmt.Sprintf("%s must start with 'mtx:' and may contain: a-z, 0-9, -, /, *, and :", field)
}

func RegexChecker(tagName string, match *regexp.Regexp, v interface{}) bool {
	st := reflect.ValueOf(v)

	switch st.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		for i := 0; i < st.Len(); i++ {
			if !RegexChecker(tagName, match, st.Index(i).Interface()) {
				return false
			}
		}

		return true
	case reflect.String:
		if st.String() == "" {
			return true
		}

		return match.MatchString(st.String())
	default:
		panic(fmt.Sprintf("invalid type for %s tag", tagName))
	}
}

func asInt(param string) int64 {
	i, err := strconv.ParseInt(param, 0, 64) //nolint:gomnd
	if err != nil {
		panic(fmt.Sprintf("cannot cast %q to int", param))
	}

	return i
}

func asUint(param string) uint64 {
	i, err := strconv.ParseUint(param, 0, 64) //nolint:gomnd
	if err != nil {
		panic(fmt.Sprintf("cannot cast %q to uint", param))
	}

	return i
}

func asFloat(param string) float64 {
	i, err := strconv.ParseFloat(param, 64) //nolint:gomnd
	if err != nil {
		panic(fmt.Sprintf("cannot cast %q to float", param))
	}

	return i
}
