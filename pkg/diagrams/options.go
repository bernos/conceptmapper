package diagrams

type D2DiagramGeneratorOption func(*D2DiagramGenerator)

func WithD2GraphModifier(mod D2GraphModifier) D2DiagramGeneratorOption {
	return func(d *D2DiagramGenerator) {
		d.graphModifiers = append(d.graphModifiers, mod)
	}
}

func WithD2RulerFactory(fn D2RulerFactory) D2DiagramGeneratorOption {
	return func(d *D2DiagramGenerator) {
		d.rulerFactory = fn
	}
}

func WithDirection(direction Direction) D2DiagramGeneratorOption {
	return func(d *D2DiagramGenerator) {
		d.direction = direction
	}
}
