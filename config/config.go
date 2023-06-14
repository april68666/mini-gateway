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
	Id          string        `yaml:"id"`
	Targets     []*Target     `yaml:"targets"`
	Discovery   string        `yaml:"discovery"`
	Protocol    string        `yaml:"protocol"`
	Timeout     int           `yaml:"timeout"`
	LoadBalance string        `yaml:"loadBalance"`
	Predicates  *Predicates   `yaml:"predicates"`
	Middlewares []*Middleware `yaml:"middlewares"`
}

type Target struct {
	Uri    string `yaml:"uri"`
	Weight int    `yaml:"weight"`
	Color  string `yaml:"color"`
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

func (e *Endpoint) Diff(endpoint *Endpoint) bool {
	if e.Id != endpoint.Id || e.Discovery != endpoint.Discovery ||
		e.Protocol != endpoint.Protocol || e.Timeout != endpoint.Timeout ||
		e.LoadBalance != endpoint.LoadBalance {
		return false
	}

	if !e.diffTarget(endpoint.Targets) {
		return false
	}

	if !e.diffPredicates(endpoint.Predicates) {
		return false
	}

	if !e.diffMiddleware(endpoint.Middlewares) {
		return false
	}

	return true
}

func (e *Endpoint) diffTarget(targets []*Target) bool {
	if len(e.Targets) != len(targets) {
		return false
	}
	for i, target := range e.Targets {
		if target.Uri != targets[i].Uri || target.Weight != targets[i].Weight || target.Color != targets[i].Color {
			return false
		}
	}
	return true
}

func (e *Endpoint) diffPredicates(predicates *Predicates) bool {
	if e.Predicates.Path != predicates.Path || e.Predicates.Method != predicates.Method {
		return false
	}

	if len(e.Predicates.Headers) != len(predicates.Headers) {
		return false
	}

	for key, value := range e.Predicates.Headers {
		if v, ok := predicates.Headers[key]; ok {
			if value != v {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (e *Endpoint) diffMiddleware(ms []*Middleware) bool {
	if len(e.Middlewares) != len(ms) {
		return false
	}

	for i, m := range e.Middlewares {
		if m.Name != ms[i].Name || m.Order != ms[i].Order {
			return false
		}

		if len(m.Args) != len(ms[i].Args) {
			return false
		}

		for key, value := range m.Args {
			if v, ok := ms[i].Args[key]; ok {
				if value != v {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}
