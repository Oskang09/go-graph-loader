package ggl

type validator interface {
	Validate(i interface{}) error
}
