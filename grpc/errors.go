//nolint:wrapcheck
package grpc

import (
	"errors"
	"fmt"

	"github.com/nielskrijger/goutils/validate"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var InternalError = status.Error(codes.Internal, "something went wrong, please try again later")

// ValidationErrors takes the validation error output and returns an
// InvalidArgument grpc error. The grpc description contains a summary,
// error details are stored as FieldViolations.
//
// Returns nil if len(errs) == 0.
func ValidationErrors(err error) error {
	if err == nil {
		return nil
	}

	var errs validate.FieldErrors
	if !errors.As(err, &errs) {
		// unexpected, interpret as internal server error
		return status.New(codes.Internal, fmt.Sprintf("unexpected error type: %s", err)).Err()
	}

	if len(errs) == 0 {
		return nil
	}

	st := status.New(codes.InvalidArgument, errs.Error())
	br := &errdetails.BadRequest{}

	for _, fieldErr := range errs {
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       fieldErr.Field,
			Description: fieldErr.Description,
		})
	}

	st, err = st.WithDetails(br)
	if err != nil {
		return status.New(codes.Internal, fmt.Sprintf("failed creating invalid argument error: %s", err)).Err()
	}

	return st.Err()
}

// ValidationError takes a field error and returns an InvalidArgument grpc error.
//
// Returns nil if err is nil.
func ValidationError(err error) error {
	if err == nil {
		return nil
	}

	var fieldErr validate.FieldError
	if !errors.As(err, &fieldErr) {
		// unexpected, interpret as internal server error
		return status.New(codes.Internal, fmt.Sprintf("unexpected error type: %s", err)).Err()
	}

	st := status.New(codes.InvalidArgument, fieldErr.Error())
	br := &errdetails.BadRequest{}
	br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
		Field:       fieldErr.Field,
		Description: fieldErr.Description,
	})

	st, err = st.WithDetails(br)
	if err != nil {
		return status.New(codes.Internal, fmt.Sprintf("failed creating invalid argument error: %s", err)).Err()
	}

	return st.Err()
}
