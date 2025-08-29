/*
Package generate exposes pre-defined parsers and generator for kickr repositories parsing to use with engine.Generate.

Example:

	func main() {
		ctx := context.Background()
		destdir, _ := os.Getwd()
		dest := filepath.Join(destdir, kickr.File)

		// read configuration
		var config kickr.Kickr
		if err := files.ReadYAML(dest, &config, os.ReadFile); err != nil {
			logger.Fatal(err)
		}
		config.EnsureDefaults()

		// run generation
		result, err := engine.Generate(ctx, destdir, &types.KickrWrapper{Kickr: config},
			[]engine.Parser[types.KickrWrapper]{generate.ParserGit, generate.ParserGolang, generate.ParserNode, generate.ParserChart},
			[]engine.Generator[types.KickrWrapper]{generate.GeneratorGitignore, generate.GeneratorLicense})
		// handle err
	}
*/
package generate
