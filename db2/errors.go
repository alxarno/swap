package db2

func DBE(text string, err error) error {
	return &DBError{errorType: text, originalError: err}
}

type DBError struct {
	errorType     string
	originalError error
}

func (e *DBError) Error() string {
	if e.originalError != nil {
		return e.errorType + " " + e.originalError.Error()
	}
	return e.errorType

}

func (e *DBError) Type() string {
	return e.errorType
}

func (e *DBError) Original() error {
	return e.originalError
}
