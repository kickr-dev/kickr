/*
Package generate exposes pre-defined parsers and generator for kickr repositories parsing to use with engine.Generate.

Example:

	func main() {
		ctx := t.Context()
		destdir, _ := os.Getwd()
		dest := filepath.Join(destdir, kickr.File)

		// read configuration
		var config kickr.Config
		if err := files.ReadYAML(dest, &config, os.ReadFile); err != nil {
			logger.Fatal(err)
		}
		config.EnsureDefaults()

		// run generation
		config, err := engine.Generate(ctx, destdir, config,
			[]engine.Parser[types.KickrGen]{generate.ParserGit, generate.ParserGolang, generate.ParserNode, generate.ParserChart},
			[]engine.Generator[types.KickrGen]{generate.GeneratorGitignore, generate.GeneratorLicense})
		// handle err
	}
*/
package generate
