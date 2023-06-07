package config

type Gateway struct {
	Port        int           `json:"port"`
	Middlewares []*Middleware `json:"middlewares"`
	Endpoints   []*Endpoint   `json:"endpoints"`
}

type Endpoint struct {
	Uris        []string      `json:"uris"`
	Protocol    string        `json:"protocol"`
	Timeout     int           `json:"timeout"`
	Predicates  *Predicates   `json:"predicates"`
	Middlewares []*Middleware `json:"middlewares"`
}

type Predicates struct {
	Path    string   `json:"path"`
	Method  string   `json:"method"`
	Headers []Header `json:"header"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Middleware struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
	Args  []Arg  `json:"args"`
}

type Arg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
