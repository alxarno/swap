package db

func DBE(text string, data error) error {
	return &errorString{s: text, originalError: data}
}

type DBError struct {
	s             string
	originalError error
}

func (e *DBError) Error() string {
	return e.s + " " + originalError.Error()
}

func (e *DBError) Type() string {
	return e.s
}

func (e *DBError) Original() error {
	return e.originalError
}
