package config

type Gateway struct {
	Http *Http `yaml:"http"`
}

type Http struct {
	Port        int           `yaml:"port"`
	Middlewares []*Middleware `yaml:"middlewares"`
	Endpoints   []*Endpoint   `yaml:"endpoints"`
}

type Endpoint struct {
	ID          string        `yaml:"id"`
	Targets     []*Target     `yaml:"targets"`
	Discovery   string        `yaml:"discovery"`
	Protocol    string        `yaml:"protocol"`
	Timeout     int           `yaml:"timeout"`
	LoadBalance string        `yaml:"loadBalance"`
	Predicates  *Predicates   `yaml:"predicates"`
	Middlewares []*Middleware `yaml:"middlewares"`
}

type Target struct {
	Uri    string            `yaml:"uri"`
	Weight int               `yaml:"weight"`
	Tags   map[string]string `yaml:"tags"`
}

type Predicates struct {
	Path    string                 `yaml:"path"`
	Method  string                 `yaml:"method"`
	Headers map[string]interface{} `yaml:"header"`
}

type Middleware struct {
	Name  string                 `yaml:"name"`
	Order int                    `yaml:"order"`
	Args  map[string]interface{} `yaml:"args"`
}
