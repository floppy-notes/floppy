package database

type Option func(dbc *DatabaseConfig)

func WithPath(path string) Option {
	return func(dbc *DatabaseConfig) {
		dbc.Path = path
	}
}
