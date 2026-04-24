package errors

const (
	CodeSuccess = 0

	CodeCommon = 10000
	CodeAuth   = 20000
	CodeUser   = 30000
	CodeFriend = 40000
	CodeMsg    = 50000
	CodeGroup  = 60000
	CodeFile   = 70000
	CodeSystem = 80000
)

type AppError struct {
	Code    int
	Message string
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) Error() string {
	return e.Message
}

var ErrInternal = New(CodeCommon, "internal error")
