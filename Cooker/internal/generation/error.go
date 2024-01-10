package generation

import "errors"

var ErrMethodNotAllow = errors.New("method not allow")
var ErrAutenticationMethodNotAllow = errors.New("autentication method not allow")
var ErrNoMaxLengthProvided = errors.New("no max length for string provided")
var ErrNoRangeProvided = errors.New("no range for int provided")
var ErrNegativeMaxLengthProvided = errors.New("max length negative provided")
var ErrInvalidRangeProvided = errors.New("max length negative provided")
var ErrNoExpectationLengthProvided = errors.New("no expectation length provided")
var ErrTypeNotAllow = errors.New("type not allow")
