package models

type Authentication struct {
	Method string `json:"method"`
	Secret string `json:"secret"`
}

type Parameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	CorrectValue interface{} `json:"correctValue"`
	MaxLength    *int        `json:"maxLength"`
	Range        *string     `json:"range"`
}
type Limiter struct {
	MaxRequests int `json:"maxRequests"`
	Seconds     int `json:"seconds"`
}
type APIConfig struct {
	Endpoint          string          `json:"endpoint"`
	Method            string          `json:"method"`
	Authentication    *Authentication `json:"authentication,omitempty"`
	Parameters        []Parameter     `json:"parameters,omitempty"`
	ExpectationLength *int            `json:"expectationLength"`
	Limiter           *Limiter        `json:"limiter"`
}
type Cooker struct {
	Port string
	Mod  int // 0 mod with frontend 1 mod without frontend
}

// NewMyStruct creates a new instance of MyStruct with default values.
func CreateCooker(port string) *Cooker {
	return &Cooker{
		Port: port,
		Mod:  0,
	}
}
func CreateCookerWithoutFrontend() *Cooker {
	return &Cooker{
		Port: "8088",
		Mod:  1,
	}
}
