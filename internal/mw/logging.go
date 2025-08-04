package mw

import (
	"context"
	"errors"
	"log"

	"PWZ1.0/internal/models/domainErrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		// если gRPC оишбка
		if s, ok := status.FromError(err); ok {
			log.Printf("error: code: %v, message: %q", s.Code(), s.Message())
			return resp, err
		}

		// другая
		mappedErr := mapErrorToStatus(err)
		log.Printf("%v", mappedErr)
		return resp, mappedErr
	}

	return resp, nil
}

func mapErrorToStatus(err error) error {
	switch {
	case errors.Is(err, domainErrors.ErrInvalidPackage),
		errors.Is(err, domainErrors.ErrValidationFailed),
		errors.Is(err, domainErrors.ErrWeightTooHeavy):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainErrors.ErrOrderAlreadyExists),
		errors.Is(err, domainErrors.ErrDuplicateOrder):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domainErrors.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domainErrors.ErrStorageExpired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domainErrors.ErrInternalError),
		errors.Is(err, domainErrors.ErrImportFailed),
		errors.Is(err, domainErrors.ErrOpenFiled),
		errors.Is(err, domainErrors.ErrReadFiled),
		errors.Is(err, domainErrors.ErrJsonFiled):
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Errorf(codes.Unknown, "неизвестная ошибка: %v", err)
	}
}
