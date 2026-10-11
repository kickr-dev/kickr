/*
Package templates provides a bunch of functions returning engine.Template struct.
Each function returns the templates of a specific topic (Chart, GitLab, Terraform, etc.),
they can be concatenated with slices.Concat.

Those functions can be used with engine.ApplyTemplate, engine.ApplyPatches or engine.GeneratorTemplates.

Note that engine.ApplyPatches is called by default in engine.ApplyTemplate.

Examples:

	func main() {
		root, _ := os.OpenRoot(".")
		defer root.Close()

		var c types.Repository
		for _, template := range templates.Chart() {
			err := engine.ApplyTemplate(templates.FS(), root, template, c)
			// handle err
		}
	}

	func main() {
		root, _ := os.OpenRoot(".")
		defer root.Close()

		var c types.Repository
		for _, template := range templates.Chart() {
			err := engine.ApplyPatches(templates.FS(), root, template, c)
			// handle err
		}
	}

	func main() {
		root, _ := os.OpenRoot(".")
		defer root.Close()
		var c types.Repository

		ctx := context.Background()
		f := engine.GeneratorTemplates(templates.FS(), templates.Chart())
		err := f(ctx, root, c)
		// handle err
	}
*/
package templates
