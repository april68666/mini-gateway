package config

type Gateway struct {
	Http *Http `json:"http"`
}

type Http struct {
	Port        int           `yaml:"port"`
	Middlewares []*Middleware `yaml:"middlewares"`
	Endpoints   []*Endpoint   `yaml:"endpoints"`
}

type Endpoint struct {
	Targets     []Target      `yaml:"targets"`
	Protocol    string        `yaml:"protocol"`
	Timeout     int           `yaml:"timeout"`
	LoadBalance string        `yaml:"load_balance"`
	Predicates  *Predicates   `yaml:"predicates"`
	Middlewares []*Middleware `yaml:"middlewares"`
}

type Target struct {
	Uri    string `yaml:"uri"`
	Weight int    `yaml:"weight"`
	Color  string `yaml:"color"`
}

type Predicates struct {
	Path    string   `yaml:"path"`
	Method  string   `yaml:"method"`
	Headers []Header `yaml:"header"`
}

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Middleware struct {
	Name  string                 `yaml:"name"`
	Order int                    `yaml:"order"`
	Args  map[string]interface{} `yaml:"args"`
}
