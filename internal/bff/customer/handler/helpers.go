package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tokyoyuan/dvd-rental/pkg/middleware"
)

// parseID extracts and parses an int32 ID from a path parameter.
func parseID(r *http.Request, param string) (int32, error) {
	s := r.PathValue(param)
	if s == "" {
		return 0, fmt.Errorf("missing %s parameter", param)
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: must be a number", param)
	}
	return int32(n), nil
}

// parsePagination extracts page_size and page from query parameters.
// Defaults: page_size=20 (max 100), page=1.
func parsePagination(r *http.Request) (pageSize, page int32) {
	pageSize, page = 20, 1

	if s := r.URL.Query().Get("page_size"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
			pageSize = int32(n)
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	if s := r.URL.Query().Get("page"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
			page = int32(n)
		}
	}

	return pageSize, page
}

// grpcToHTTPError maps a gRPC error to an HTTP JSON error response.
func grpcToHTTPError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
		return
	}

	var httpStatus int
	var code string

	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus, code = http.StatusBadRequest, "INVALID_ARGUMENT"
	case codes.NotFound:
		httpStatus, code = http.StatusNotFound, "NOT_FOUND"
	case codes.AlreadyExists:
		httpStatus, code = http.StatusConflict, "ALREADY_EXISTS"
	case codes.PermissionDenied:
		httpStatus, code = http.StatusForbidden, "PERMISSION_DENIED"
	case codes.Unauthenticated:
		httpStatus, code = http.StatusUnauthorized, "UNAUTHENTICATED"
	case codes.FailedPrecondition:
		httpStatus, code = http.StatusConflict, "FAILED_PRECONDITION"
	case codes.Unavailable:
		httpStatus, code = http.StatusServiceUnavailable, "UNAVAILABLE"
	case codes.DeadlineExceeded:
		httpStatus, code = http.StatusGatewayTimeout, "DEADLINE_EXCEEDED"
	default:
		httpStatus, code = http.StatusInternalServerError, "INTERNAL_ERROR"
	}

	middleware.WriteJSONError(w, httpStatus, code, st.Message())
}

// timestampToString converts a proto Timestamp to an RFC3339 string.
// Returns "" for nil timestamps.
func timestampToString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}

// readJSON decodes a JSON request body into v.
func readJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}
