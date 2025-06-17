package mw

import (
	"context"
	"encoding/json"
	"net/http"

	"PWZ1.0/internal/models/domainErrors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

type customError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func CustomErrorHandler(_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	codeStr := "UNKNOWN"

	for originalErr, customCode := range domainErrors.ErrorCodes {
		if st.Message() == originalErr.Error() {
			codeStr = customCode
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(runtime.HTTPStatusFromCode(st.Code()))

	resp := customError{}
	resp.Error.Code = codeStr
	resp.Error.Message = st.Message()

	_ = json.NewEncoder(w).Encode(resp)
}
