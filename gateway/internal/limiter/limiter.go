package limiter

import (
	"golang.org/x/time/rate"
)

var GlobalLimiter = rate.NewLimiter(rate.Every(TIME), TOKEN) //10 richieste al secondo
