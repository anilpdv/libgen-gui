package network

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"syscall"
)

// ErrorKind categorizes network failures for UX and retry logic.
type ErrorKind int

const (
	ErrorUnknown ErrorKind = iota
	ErrorNetwork
	ErrorTimeout
	ErrorHTTP
	ErrorInvalidResponse
	ErrorCancelled
)

func (k ErrorKind) String() string {
	switch k {
	case ErrorNetwork:
		return "network_error"
	case ErrorTimeout:
		return "timeout"
	case ErrorHTTP:
		return "http_error"
	case ErrorInvalidResponse:
		return "invalid_response"
	case ErrorCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// NetworkError wraps an underlying network failure with typed taxonomy.
type NetworkError struct {
	Kind       ErrorKind
	StatusCode int
	URL        string
	Err        error
}

func (e *NetworkError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%s] HTTP %d on %s: %v", e.Kind, e.StatusCode, e.URL, e.Err)
	}
	if e.URL != "" {
		return fmt.Sprintf("[%s] on %s: %v", e.Kind, e.URL, e.Err)
	}
	return fmt.Sprintf("[%s]: %v", e.Kind, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// UserFriendlyMessage returns a clear, actionable message for the UI without developer jargon.
func (e *NetworkError) UserFriendlyMessage() string {
	if e == nil {
		return ""
	}
	switch e.Kind {
	case ErrorTimeout:
		return "Connection timed out. The server took too long to respond."
	case ErrorNetwork:
		return "Unable to connect to the mirror server. Please check your internet connection."
	case ErrorCancelled:
		return "Operation was cancelled."
	case ErrorInvalidResponse:
		return "Server returned an unexpected or malformed response."
	case ErrorHTTP:
		switch {
		case e.StatusCode == http.StatusBadGateway:
			return "Mirror server error (502 Bad Gateway). Please try another mirror."
		case e.StatusCode == http.StatusServiceUnavailable:
			return "Mirror is temporarily unavailable (503). Retrying may help."
		case e.StatusCode == http.StatusGatewayTimeout:
			return "Gateway timed out (504). Mirror is responding slowly."
		case e.StatusCode == http.StatusForbidden:
			return "Mirror access was restricted (403)."
		case e.StatusCode == http.StatusNotFound:
			return "Requested resource was not found on this mirror (404)."
		case e.StatusCode >= 500:
			return fmt.Sprintf("Mirror internal server error (HTTP %d).", e.StatusCode)
		case e.StatusCode >= 400:
			return fmt.Sprintf("Server rejected request (HTTP %d).", e.StatusCode)
		}
	}
	return "Network request failed. Please try again."
}

// ClassifyError inspects an arbitrary error and wraps it in a *NetworkError.
func ClassifyError(err error, rawURL string, statusCode int) *NetworkError {
	if err == nil && statusCode >= 200 && statusCode < 400 {
		return nil
	}

	netErr := &NetworkError{
		StatusCode: statusCode,
		URL:        rawURL,
		Err:        err,
	}

	if statusCode >= 400 {
		netErr.Kind = ErrorHTTP
		if err == nil {
			netErr.Err = fmt.Errorf("HTTP status %d", statusCode)
		}
		return netErr
	}

	if err == nil {
		netErr.Kind = ErrorUnknown
		netErr.Err = errors.New("unknown error")
		return netErr
	}

	var existing *NetworkError
	if errors.As(err, &existing) {
		return existing
	}

	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "context canceled") || strings.Contains(errMsg, "operation was canceled"):
		netErr.Kind = ErrorCancelled
	case strings.Contains(errMsg, "deadline exceeded") || strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "timed out"):
		netErr.Kind = ErrorTimeout
	case strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "no route to host") ||
		strings.Contains(errMsg, "reset by peer") ||
		strings.Contains(errMsg, "lookup ") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "handshake failure") ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET):
		netErr.Kind = ErrorNetwork
	case strings.Contains(errMsg, "invalid response") || strings.Contains(errMsg, "unexpected json") || strings.Contains(errMsg, "html error page"):
		netErr.Kind = ErrorInvalidResponse
	default:
		var netOpErr *net.OpError
		if errors.As(err, &netOpErr) {
			if netOpErr.Timeout() {
				netErr.Kind = ErrorTimeout
			} else {
				netErr.Kind = ErrorNetwork
			}
		} else {
			netErr.Kind = ErrorNetwork
		}
	}

	return netErr
}
