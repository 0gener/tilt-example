package components

// Option represents a component configuration option.
type Option func(Component) error

// Options applies the provided list of options.
func Options(opts ...Option) Option {
	return func(s Component) error {
		for _, opt := range opts {
			if err := opt(s); err != nil {
				return err
			}
		}

		return nil
	}
}
