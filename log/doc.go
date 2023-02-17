// This package provide logger abstraction and default implementation.
// This package also provide global logger which use the default implementation by defaults.
// Libraries might use the global logger to log their internal event or debuging purpose, libraries
// that use global logger should use Debug level to avoid cluter the user logs.
// SimpleLogger which is the default implementation should be replaced by more reliable logging
// library implementation such as `zaplog`, you can set global logger by calling:
// log.SetLogger(logger)
package log
