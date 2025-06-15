package common

type RequestValidationError struct {
	Massage string
}

func (err *RequestValidationError) Error() string {
	return err.Massage
}

type AlreadyExistsError struct {
	Massage string
}

func (err *AlreadyExistsError) Error() string {
	return err.Massage
}

type NotFoundError struct {
	Massage string
}

func (err *NotFoundError) Error() string {
	return err.Massage
}
