package repositories

const (
	ErrCodeValidationError = iota
	ErrInternal
)

var _ error = &RepositoryError{}

const activeClause = "expires_at IS NULL OR expires_at > ?"

type RepositoryError struct {
	ErrCode int
	Message string
}

func (err *RepositoryError) Error() string {
	return err.Message
}
