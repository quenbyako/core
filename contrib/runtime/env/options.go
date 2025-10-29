package env

// parseParams for the parser.
type parseParams struct {
	environment map[string]string
	prefix      string
	onSet       func(tag string, value any, isDefault bool)
}

type Option func(*parseParams)

func WithEnvironment(env map[string]string) Option {
	return func(p *parseParams) { p.environment = env }
}

func WithPrefix(prefix string) Option {
	return func(p *parseParams) { p.prefix = prefix }
}

func WithOnSet(onSet func(tag string, value any, isDefault bool)) Option {
	return func(p *parseParams) { p.onSet = onSet }
}

func buildParseParams(opts ...Option) (parseParams, error) {
	p := parseParams{
		environment: nil,
	}
	for _, opt := range opts {
		opt(&p)
	}

	if err := p.validate(); err != nil {
		return parseParams{}, err
	}

	return p, nil
}

func (p *parseParams) validate() error {
	return nil
}

func (p *parseParams) keyWithPrefix(key string) string {
	return p.prefix + key
}

func (p *parseParams) getEnv(key string) (string, bool) {
	val, ok := p.environment[p.prefix+key]
	return val, ok
}
