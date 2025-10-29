package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcstatus interface {
	error
	GRPCStatus() *status.Status
}

type invalidArgumentError struct {
	Violations []*errdetails.BadRequest_FieldViolation
}

var _ grpcstatus = (*invalidArgumentError)(nil) //nolint:errcheck // WHAT???

func errInvalidArgument(violations ...*errdetails.BadRequest_FieldViolation) *invalidArgumentError {
	filtered := make([]*errdetails.BadRequest_FieldViolation, 0, len(violations))

	for _, v := range violations {
		if v != nil {
			filtered = append(filtered, v)
		}
	}

	return &invalidArgumentError{
		Violations: filtered,
	}
}

func (e *invalidArgumentError) Error() string {
	if len(e.Violations) == 0 {
		return "invalid argument"
	}

	msg := "invalid argument: "

	for i, v := range e.Violations {
		if i > 0 {
			msg += ", "
		}

		msg += v.GetField() + ": " + v.GetDescription()
	}

	return msg
}

func (e *invalidArgumentError) GRPCStatus() *status.Status {
	s := status.New(codes.InvalidArgument, e.Error())
	if len(e.Violations) == 0 {
		return s
	}

	return must(s.WithDetails(
		&errdetails.ErrorInfo{
			Reason:   "REASON_INVALID_ARGUMENT",
			Domain:   "",
			Metadata: nil,
		},
		&errdetails.BadRequest{
			FieldViolations: e.Violations,
		},
	))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err) //nolint:forbidigo // there is a purpose for this...
	}

	return v
}
