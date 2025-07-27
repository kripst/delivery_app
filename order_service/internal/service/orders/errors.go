package orders

import "errors"

var (
	ErrNilStruct = errors.New("error nil struct")
	ErrNothingChangedStruct = errors.New("error nothing changed")
)
