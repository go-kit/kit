package log

// Levels provides a default set of leveled loggers.
type Levels struct {
	Debug Logger
	Info  Logger
	Error Logger
}

type levelOptions struct {
	levelKey   string
	debugValue string
	infoValue  string
	errorValue string
}

// LevelOption sets a parameter for leveled loggers.
type LevelOption func(*levelOptions)

// LevelKey sets the key for the field used to indicate log level. By default,
// the key is "level".
func LevelKey(key string) LevelOption {
	return func(o *levelOptions) { o.levelKey = key }
}

// DebugLevelValue sets the value for the field used to indicate the debug log
// level. By default, the value is "DEBUG".
func DebugLevelValue(value string) LevelOption {
	return func(o *levelOptions) { o.debugValue = value }
}

// InfoLevelValue sets the value for the field used to indicate the debug log
// level. By default, the value is "INFO".
func InfoLevelValue(value string) LevelOption {
	return func(o *levelOptions) { o.infoValue = value }
}

// ErrorLevelValue sets the value for the field used to indicate the debug log
// level. By default, the value is "ERROR".
func ErrorLevelValue(value string) LevelOption {
	return func(o *levelOptions) { o.errorValue = value }
}

// NewLevels returns a new set of leveled loggers based on the base logger.
func NewLevels(base Logger, options ...LevelOption) Levels {
	opts := &levelOptions{
		levelKey:   "level",
		debugValue: "DEBUG",
		infoValue:  "INFO",
		errorValue: "ERROR",
	}
	for _, option := range options {
		option(opts)
	}
	return Levels{
		Debug: With(base, opts.levelKey, opts.debugValue),
		Info:  With(base, opts.levelKey, opts.infoValue),
		Error: With(base, opts.levelKey, opts.errorValue),
	}
}
