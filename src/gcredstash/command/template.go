package command

import (
	"bytes"
	"fmt"
	"gcredstash"
	"html/template"
	"os"
	"strings"
)

type TemplateCommand struct {
	Meta
}

func (c *TemplateCommand) parseArgs(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("too few arguments")
	}

	if len(args) > 1 {
		return "", fmt.Errorf("too few arguments")
	}

	tmplFile := args[0]

	return tmplFile, nil
}

func (c *TemplateCommand) readTemplate(filename string) (string, error) {
	var content string

	if filename == "-" {
		content = gcredstash.ReadStdin()
	} else {
		var err error
		content, err = gcredstash.ReadFile(filename)

		if err != nil {
			return "", nil
		}
	}

	return content, nil
}

func (c *TemplateCommand) getCredential(credential string, context map[string]string) (string, error) {
	value, err := c.Driver.GetSecret(credential, "", c.Table, context)

	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *TemplateCommand) executeTemplate(name string, content string) (string, error) {
	tmpl := template.New(name)

	tmpl = tmpl.Funcs(template.FuncMap{
		"get": func(args ...interface{}) string {
			newArgs := []string{}

			if len(args) < 1 {
				return "(too few arguments)"
			}

			for _, arg := range args {
				str, ok := arg.(string)

				if !ok {
					return fmt.Sprintf("(invalid string: %v)", arg)
				}

				newArgs = append(newArgs, str)
			}

			credential := newArgs[0]
			context, err := gcredstash.ParseContext(newArgs[1:])

			if err != nil {
				return err.Error()
			}

			value, err := c.getCredential(credential, context)

			if err != nil {
				return err.Error()
			}

			return value
		},
	})

	tmpl, err := tmpl.Parse(content)

	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	tmpl.Execute(buf, nil)

	return buf.String(), nil
}

func (c *TemplateCommand) RunImpl(args []string) (string, error) {
	tmplFile, err := c.parseArgs(args)

	if err != nil {
		return "", err
	}

	tmplContent, err := c.readTemplate(tmplFile)

	if err != nil {
		return "", err
	}

	out, err := c.executeTemplate(tmplFile, tmplContent)

	return out, err
}

func (c *TemplateCommand) Run(args []string) int {
	out, err := c.RunImpl(args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		return 1
	}

	fmt.Print(out)

	return 0
}

func (c *TemplateCommand) Synopsis() string {
	return "Parse a template file with credentials"
}

func (c *TemplateCommand) Help() string {
	helpText := `
usage: gcredstash template template_file
`
	return strings.TrimSpace(helpText)
}
