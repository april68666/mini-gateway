package config

type Gateway struct {
	Port        int           `json:"port"`
	Middlewares []*Middleware `json:"middlewares"`
	Endpoints   []*Endpoint   `json:"endpoints"`
}

type Endpoint struct {
	Targets     []Target      `json:"targets"`
	Protocol    string        `json:"protocol"`
	Timeout     int           `json:"timeout"`
	LoadBalance string        `json:"load_balance"`
	Predicates  *Predicates   `json:"predicates"`
	Middlewares []*Middleware `json:"middlewares"`
}

type Target struct {
	Uri    string `json:"uri"`
	Weight int    `json:"weight"`
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
