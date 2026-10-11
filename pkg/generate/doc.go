/*
Package generate exposes pre-defined parsers and generators for kickr repositories to use with engine.Generate.

Example:

	func main() {
		ctx := context.Background()
		destdir, _ := os.Getwd()
		filename := kickr.File(destdir)

		// read configuration
		var config kickr.Kickr
		if err := files.ReadYAML(os.DirFS(destdir), filename, &config); err != nil {
			logger.Fatal(err.Error())
		}
		config.EnsureDefaults()

		client := &http.Client{Transport: engine.NewLoggerURL(http.DefaultTransport)}

		// run generation
		err := engine.Generate(ctx, destdir, types.Repository{Config: config},
			[]engine.Parser[types.Repository]{generate.ParserGit, generate.ParserModules, generate.ParserGolang(client), generate.ParserNode, generate.ParserHelm},
			[]engine.Generator[types.Repository]{generate.GeneratorGitignore(client), generate.GeneratorLicense(client)})
		// handle err
	}
*/
package generate
