package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"gopkg.in/yaml.v3"
)

var Version = "dev"

func main() {
	var (
		valuesFiles   multiStringFlag
		templateFiles multiStringFlag
		templateDir   string
		outputDir     string
	)

	flag.Var(&valuesFiles, "value", "Values YAML file(s), can be specified multiple times")
	flag.Var(&templateFiles, "template", "Template YAML file(s), can be specified multiple times")
	flag.StringVar(&templateDir, "template-dir", "", "Directory containing template files")
	flag.StringVar(&outputDir, "output-dir", "", "Directory to write output files")
	setValues := flag.String("set", "", "Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	outputFile := flag.String("output", "", "Output file (when processing a single template)")
	delimiter := flag.String("delimiter", "{{,}}", "Template delimiter")
	strict := flag.Bool("strict", false, "Strict mode (missing keys will cause an error)")
	flag.Parse()

	fmt.Println("# Helmlet version:", Version)
	
	// Check that at least one template source is specified
	if len(templateFiles) == 0 && templateDir == "" {
		fmt.Println("Usage: helmlet --value values.yaml --template template.yaml [--set key1=val1,key2=val2] [--output output.yaml]")
		fmt.Println("       helmlet --value values.yaml --template-dir templates/ [--output-dir output/]")
		return
	}

	// Load and merge values from all files
	values := make(map[string]interface{})
	for _, valuesFile := range valuesFiles {
		if err := loadValuesFile(valuesFile, values); err != nil {
			fmt.Println("Error with values file:", err)
			return
		}
	}

	// Parse and apply set values
	if *setValues != "" {
		setValuesMap := parseSetValues(*setValues)
		for k, v := range setValuesMap {
			setNestedValue(values, strings.Split(k, "."), v)
		}
	}

	// Set up function map for templates
	funcMap := createFuncMap()
	delimiters := strings.Split(*delimiter, ",")
	
	// Process template files
	templates := make([]string, 0)
	
	// Add individual template files
	templates = append(templates, templateFiles...)
	
	// Add templates from directory if specified
	if templateDir != "" {
		dirTemplates, err := findTemplatesInDir(templateDir)
		if err != nil {
			fmt.Println("Error reading template directory:", err)
			return
		}
		templates = append(templates, dirTemplates...)
	}
	
	// Process all templates
	if len(templates) == 0 {
		fmt.Println("No templates found to process")
		return
	}
	
	// Create output directory if needed
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Println("Error creating output directory:", err)
			return
		}
	}
	
	// Special case: single template with outputFile
	if len(templates) == 1 && *outputFile != "" {
		processTemplate(templates[0], *outputFile, values, funcMap, delimiters, *strict)
		return
	}
	
	// Process multiple templates
	for _, templatePath := range templates {
		// Determine output file name
		var outFile string
		if outputDir != "" {
			outFile = filepath.Join(outputDir, filepath.Base(templatePath))
		} else if *outputFile != "" {
			outFile = *outputFile
		} else {
			// Output to stdout with a header showing which template
			processTemplateToStdout(templatePath, values, funcMap, delimiters, *strict)
			continue
		}
		
		processTemplate(templatePath, outFile, values, funcMap, delimiters, *strict)
	}
}

// multiStringFlag allows for repeated flag values
type multiStringFlag []string

func (m *multiStringFlag) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func loadValuesFile(filename string, values map[string]interface{}) error {
	valuesData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", filename, err)
	}
	
	// Parse YAML into a new map
	newValues := make(map[string]interface{})
	if err := yaml.Unmarshal(valuesData, &newValues); err != nil {
		return fmt.Errorf("parsing YAML from %s: %w", filename, err)
	}
	
	// Merge with existing values (values from later files override earlier ones)
	mergeValues(values, newValues)
	return nil
}

func mergeValues(dest, src map[string]interface{}) {
	for k, v := range src {
		// If both are maps, merge them recursively
		if destMap, ok := dest[k].(map[string]interface{}); ok {
			if srcMap, ok := v.(map[string]interface{}); ok {
				mergeValues(destMap, srcMap)
				continue
			}
		}
		// Otherwise, src overwrites dest
		dest[k] = v
	}
}

func findTemplatesInDir(dir string) ([]string, error) {
	var templates []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".yaml") || 
						   strings.HasSuffix(path, ".yml") ||
						   strings.HasSuffix(path, ".tpl")) {
			templates = append(templates, path)
		}
		return nil
	})
	return templates, err
}

func createFuncMap() template.FuncMap {
	return template.FuncMap{
		"tpl": func(tplString string, data interface{}) (string, error) {
			var buf bytes.Buffer
			tmpl, err := template.New("tpl").Parse(tplString)
			if err != nil {
				return "", err
			}
			err = tmpl.Execute(&buf, data)
			return buf.String(), err
		},
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil {
				return defaultValue
			}
			return value
		},
		"quote": func(s interface{}) string {
			switch v := s.(type) {
			case string:
				return strconv.Quote(v)
			default:
				return strconv.Quote(fmt.Sprintf("%v", v))
			}
		},
		"FilesGet": func(file string) (string, error) {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Sprintf("# %s", err), nil // suppress error but print it
			}
			return string(data), nil
		},
		"indent": func(spaces int, v string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.Replace(v, "\n", "\n"+pad, -1)
		},
	}
}

func processTemplate(templatePath, outputPath string, values map[string]interface{}, funcMap template.FuncMap, delimiters []string, strict bool) {
	templateData, err := readUTF8File(templatePath)
	if err != nil {
		fmt.Printf("# Error reading template file %s: %v\n", templatePath, err)
		return
	}

	tmpl, err := template.New(filepath.Base(templatePath)).
		Delims(delimiters[0], delimiters[1]).
		Funcs(funcMap).
		Parse(string(templateData))

	if strict {
		tmpl.Option("missingkey=error")
	}

	if err != nil {
		fmt.Printf("# Error parsing template %s: %v\n", templatePath, err)
		return
	}

	data := map[string]interface{}{
		"Values": values,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		fmt.Printf("# Error executing template %s: %v\n", templatePath, err)
		return
	}

	utf8Data, _, err := transform.Bytes(unicode.UTF8.NewEncoder(), buf.Bytes())
	if err != nil {
		fmt.Printf("# Error converting output to UTF-8 for %s: %v\n", templatePath, err)
		return
	}

	if outputPath != "" {
		err = os.WriteFile(outputPath, utf8Data, 0644)
		if err != nil {
			fmt.Printf("# Error writing to output file %s: %v\n", outputPath, err)
			return
		}
		fmt.Printf("# Output written to %s\n", outputPath)
	} else {
		os.Stdout.Write(utf8Data)
	}
}

func processTemplateToStdout(templatePath string, values map[string]interface{}, funcMap template.FuncMap, delimiters []string, strict bool) {
	fmt.Printf("\n# Processing template: %s\n", templatePath)
	fmt.Println("# -----------------------------------------------")
	processTemplate(templatePath, "", values, funcMap, delimiters, strict)
}

func readUTF8File(filename string) ([]byte, error) {
	rawData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if !utf8.Valid(rawData) {
		utf8Data, _, err := transform.Bytes(unicode.UTF8.NewDecoder(), rawData)
		fmt.Printf("# Converted %s to UTF-8\n", filename)
		if err != nil {
			return nil, err
		}
		return utf8Data, nil
	}

	return rawData, nil
}

func parseSetValues(setValues string) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Use CSV parsing to handle commas within quoted values
	r := csv.NewReader(strings.NewReader(setValues))
	r.Comma = ','
	r.LazyQuotes = true
	
	records, err := r.ReadAll()
	if err != nil {
		// Fallback to simple splitting if CSV parsing fails
		pairs := strings.Split(setValues, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				result[kv[0]] = kv[1]
			}
		}
		return result
	}
	
	// Process the CSV records
	for _, record := range records {
		for _, item := range record {
			kv := strings.SplitN(item, "=", 2)
			if len(kv) == 2 {
				result[kv[0]] = kv[1]
			}
		}
	}
	
	return result
}

func setNestedValue(m map[string]interface{}, keys []string, value interface{}) {
	if len(keys) == 1 {
		m[keys[0]] = value
		return
	}
	if m[keys[0]] == nil {
		m[keys[0]] = make(map[string]interface{})
	}
	nextMap, ok := m[keys[0]].(map[string]interface{})
	if !ok {
		nextMap = make(map[string]interface{})
		for k, v := range m[keys[0]].(map[interface{}]interface{}) {
			nextMap[k.(string)] = v
			fmt.Println(v)
		}
		m[keys[0]] = nextMap
	}
	setNestedValue(nextMap, keys[1:], value)
}