/*
Package initialize exposes pre-defined groups for kickr configuration initialization with engine.Initialize.

Example:

	func main() {
		config, err := engine.Initialize(ctx, engine.WithFormGroups(initialize.Maintainer, initialize.License, initialize.Defaults))
		// handle err
	}
*/
package initialize
