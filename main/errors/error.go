package errors

const (
	Query_ERROR        = -1000
	Invalid_NAME_ERROR = -1001
)

type APIError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var (
	QueryError = &APIError{
		Query_ERROR,
		"Query error",
	}

	InvalidName = &APIError{
		Invalid_NAME_ERROR,
		"Invalid Name",
	}
)
