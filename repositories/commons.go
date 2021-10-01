package repositories

const (
	ErrCodeValidationError = iota
	ErrInternal
)

var _ error = &RepositoryError{}

type RepositoryError struct {
	ErrCode int
	Message string
}

func (err *RepositoryError) Error() string {
	return err.Message
}
