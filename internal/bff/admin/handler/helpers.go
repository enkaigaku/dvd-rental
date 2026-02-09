package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

// grpcToHTTPStatus converts a gRPC status code to HTTP status code.
func grpcToHTTPStatus(err error) int {
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch st.Code() {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	default:
		return http.StatusInternalServerError
	}
}

// handleGRPCError writes the appropriate HTTP error based on gRPC error.
func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeError(w, grpcToHTTPStatus(err), st.Message())
}

// parseIntParam parses an integer from a path parameter.
func parseIntParam(r *http.Request, name string) (int32, error) {
	s := r.PathValue(name)
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

// parsePagination extracts page_size and page from query parameters.
func parsePagination(r *http.Request) (pageSize, page int32) {
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.ParseInt(ps, 10, 32); err == nil && v > 0 {
			pageSize = int32(v)
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.ParseInt(p, 10, 32); err == nil && v > 0 {
			page = int32(v)
		}
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if page == 0 {
		page = 1
	}
	return pageSize, page
}

// parseQueryInt32 parses an optional int32 query parameter.
func parseQueryInt32(r *http.Request, name string) int32 {
	s := r.URL.Query().Get(name)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(v)
}

// parseQueryBool parses an optional bool query parameter.
func parseQueryBool(r *http.Request, name string) bool {
	s := r.URL.Query().Get(name)
	return s == "true" || s == "1"
}

// decodeJSON decodes JSON request body into the given target.
func decodeJSON(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}
