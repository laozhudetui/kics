package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/Checkmarx/kics/pkg/parser/terraform/converter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	// hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/pkg/errors"
)

// RetriesDefaultValue is default number of times a parser will retry to execute
const RetriesDefaultValue = 50

// Converter returns content json, error line, error
type Converter func(file *hcl.File) (model.Document, int, error)

// Parser struct that contains the function to parse file and the number of retries if something goes wrong
type Parser struct {
	convertFunc  Converter
	numOfRetries int
}

type tfvarsSource struct {
	value  cty.Value
	source byte
}

type inputVariables map[string]tfvarsSource

// NewDefault initializes a parser with Parser default values
func NewDefault() *Parser {
	// here comes tf-var and tf-var-files and create Args variables
	return &Parser{
		numOfRetries: RetriesDefaultValue,
		convertFunc:  converter.DefaultConverted,
	}
}

// Resolve - replace or modifies in-memory content before parsing
func (p *Parser) Resolve(fileContent []byte, filename string) (*[]byte, error) {
	valueMap, _ := getInputVariables(filepath.Dir(filename))
	validIdentifier := `var\.[0-9a-zA-Z\-_]+(\[[0-9]+\]|\[\"[0-9a-zA-Z\-_]+\"\])*`
	fullRegexRule := fmt.Sprintf(`(%s)|(\$\{%s\})`, validIdentifier, validIdentifier)
	regex := regexp.MustCompile(fullRegexRule)

	identifiers := regex.FindAll(fileContent, -1)
	for _, identifier := range identifiers {
		keyIdentifier := strings.Replace(string(identifier), "var.", "", -1)
		if value, ok := valueMap[keyIdentifier]; !ok {
			fmt.Printf("error")
		} else {
			fmt.Printf("%s", value)
		}
		fmt.Printf("%q", identifier)
	}
	return &fileContent, nil
}

func getInputVariables(currentPath string) (inputVariables, error) {
	terraformFilepath := filepath.Join(currentPath, "terraform.tfvars")
	file, err := os.ReadFile(terraformFilepath)
	if err != nil {
		return nil, err
	}
	var f *hcl.File
	if strings.HasSuffix(terraformFilepath, ".json") {
		// f, _ = hcljson.Parse(file, terraformFilepath)
		if f == nil || f.Body == nil {
			return nil, nil
		}
	} else {
		f, _ = hclsyntax.ParseConfig(file, terraformFilepath, hcl.Pos{Line: 1, Column: 1})
		if f == nil || f.Body == nil {
			return nil, nil
		}
	}

	err = checkTfvarsValid(f, terraformFilepath)
	if err != nil {
		return nil, err
	}

	attrs, _ := f.Body.JustAttributes()
	iVariables := make(inputVariables, 0)
	//Here need to change what to get
	for name, attr := range attrs {
		val, _ := attr.Expr.Value(nil)
		iVariables[name] = tfvarsSource{
			value:  val,
			source: 't',
		}
	}
	return iVariables, nil
}

func checkTfvarsValid(f *hcl.File, filename string) error {
	content, _, _ := f.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	})
	if len(content.Blocks) > 0 {
		return fmt.Errorf("failed to get variables from %s, .tfvars file is used to assing values not to declare new variables.", filename)
	}
	return nil
}

// Parse execute parser for the content in a file
func (p *Parser) Parse(path string, content []byte) ([]model.Document, error) {
	var (
		fc        model.Document
		lineOfErr int
		parseErr  error
	)

	for try := 0; try < p.numOfRetries; try++ {
		fc, lineOfErr, parseErr = p.doParse(content, filepath.Base(path))
		if parseErr != nil && lineOfErr != 0 {
			content = p.removeProblematicLine(content, lineOfErr)
			continue
		}

		break
	}

	return []model.Document{fc}, errors.Wrap(parseErr, "failed terraform parse")
}

// SupportedExtensions returns Terraform extensions
func (p *Parser) SupportedExtensions() []string {
	return []string{".tf"}
}

// SupportedTypes returns types supported by this parser, which are terraform
func (p *Parser) SupportedTypes() []string {
	return []string{"Terraform"}
}

// GetKind returns Terraform kind parser
func (p *Parser) GetKind() model.FileKind {
	return model.KindTerraform
}

func (p *Parser) removeProblematicLine(content []byte, line int) []byte {
	lines := strings.Split(string(content), "\n")
	if line > 0 && line <= len(lines) {
		lines[line-1] = ""
		return []byte(strings.Join(lines, "\n"))
	}
	return content
}

func (p *Parser) doParse(content []byte, fileName string) (json model.Document, errLine int, err error) {
	file, diagnostics := hclsyntax.ParseConfig(content, fileName, hcl.Pos{Byte: 0, Line: 1, Column: 1})

	if diagnostics != nil && diagnostics.HasErrors() && len(diagnostics.Errs()) > 0 {
		err := diagnostics.Errs()[0]
		line := 0

		if e, ok := err.(*hcl.Diagnostic); ok {
			line = e.Subject.Start.Line
		}

		return nil, line, errors.Wrap(err, "failed to parse file")
	}

	return p.convertFunc(file)
}
