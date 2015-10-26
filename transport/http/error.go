package http

// BadRequestError is an error in decoding the request.
type BadRequestError struct{ error }

// Raw returns the raw embedded error
func (err BadRequestError) Raw() error {
	return err.error
}
