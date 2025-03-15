package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if err := generateFile(gen, f); err != nil {
				return err
			}
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}
	filenameWithoutExt := strings.TrimSuffix(string(*file.Proto.Name), ".proto")
	// Build the filename where generated code will be merged
	filename := "./pkg/services/" + filenameWithoutExt + "_server.go"

	// Generate new content using a string builder
	var sb strings.Builder
	sb.WriteString("package services\n\n")
	sb.WriteString("import (\n")
	sb.WriteString(` . "generated"` + "\n")
	sb.WriteString(`    "context"` + "\n")
	sb.WriteString(`    "db"` + "\n")
	sb.WriteString(`    "go.uber.org/zap"` + "\n")
	sb.WriteString(")\n\n")

	// Generate service implementations for each service in the proto file
	for _, service := range file.Services {
		generateServiceImplementation(&sb, service)
	}

	newContent := sb.String()

	// Merge the new content with any existing file content
	mergedContent, err := mergeWithExisting(filename, newContent)
	if err != nil {
		return err
	}

	// Write the merged content to the file
	return os.WriteFile(filename, []byte(mergedContent), 0644)
}

func generateServiceImplementation(sb *strings.Builder, service *protogen.Service) {
	structName := service.GoName + "ServiceServer"

	// Generate the service struct
	sb.WriteString("type " + structName + " struct {\n")
	sb.WriteString("    Unimplemented" + service.GoName + "Server\n")
	sb.WriteString("    PrismaClient *db.PrismaClient\n")
	sb.WriteString("    Logger       *zap.SugaredLogger\n")
	sb.WriteString("}\n\n")

	// Generate a constructor
	sb.WriteString("func New" + structName + "() *" + structName + " {\n")
	sb.WriteString("    return &" + structName + "{}\n")
	sb.WriteString("}\n\n")

	// Generate methods for each RPC
	for _, method := range service.Methods {
		generateMethodImplementation(sb, method, structName)
	}
}

func generateMethodImplementation(sb *strings.Builder, method *protogen.Method, structName string) {
	methodName := method.GoName
	inputType := method.Input.GoIdent.GoName
	outputType := method.Output.GoIdent.GoName

	signature := fmt.Sprintf("func (s *%s) %s(ctx context.Context, req *%s) (*%s, error) {",
		structName, methodName, inputType, outputType)
	sb.WriteString(signature + "\n")
	sb.WriteString("    // TODO: Implement " + methodName + "\n")
	sb.WriteString("    return &" + outputType + "{}, nil\n")
	sb.WriteString("}\n\n")
}

// mergeWithExisting reads an existing file (if any) and merges it with newContent,
// avoiding duplicate method definitions.
func mergeWithExisting(filename, newContent string) (string, error) {
	existing, err := os.ReadFile(filename)
	if err != nil {
		// If the file doesn't exist, just return the new content.
		return newContent, nil
	}
	merged := mergeContent(string(existing), newContent)
	return merged, nil
}

// mergeContent uses a simple strategy: it identifies method definitions by looking for
// "func (s *..." signatures and then only appends methods that are not already present.
func mergeContent(existing, new string) string {
	// Regular expression to capture service struct and method names.
	methodRegex := regexp.MustCompile(`func \(s \*([A-Za-z0-9_]+)\) ([A-Za-z0-9_]+)\(`)
	existingMethods := make(map[string]bool)
	for _, match := range methodRegex.FindAllStringSubmatch(existing, -1) {
		if len(match) >= 3 {
			key := match[1] + "::" + match[2]
			existingMethods[key] = true
		}
	}

	// Break new content into lines and extract method blocks.
	lines := strings.Split(new, "\n")
	var methodsToAdd []string
	var currentMethod []string
	inMethod := false
	for _, line := range lines {
		if strings.HasPrefix(line, "func (s *") {
			if inMethod && len(currentMethod) > 0 {
				methodBlock := strings.Join(currentMethod, "\n")
				m := methodRegex.FindStringSubmatch(currentMethod[0])
				if len(m) >= 3 {
					key := m[1] + "::" + m[2]
					if !existingMethods[key] {
						methodsToAdd = append(methodsToAdd, methodBlock)
					}
				}
			}
			inMethod = true
			currentMethod = []string{line}
		} else if inMethod {
			currentMethod = append(currentMethod, line)
			if line == "" {
				// End of a method block
				methodBlock := strings.Join(currentMethod, "\n")
				m := methodRegex.FindStringSubmatch(currentMethod[0])
				if len(m) >= 3 {
					key := m[1] + "::" + m[2]
					if !existingMethods[key] {
						methodsToAdd = append(methodsToAdd, methodBlock)
					}
				}
				inMethod = false
				currentMethod = nil
			}
		}
	}
	// In case a method was being collected but not closed by an empty line
	if inMethod && len(currentMethod) > 0 {
		m := methodRegex.FindStringSubmatch(currentMethod[0])
		if len(m) >= 3 {
			key := m[1] + "::" + m[2]
			if !existingMethods[key] {
				methodsToAdd = append(methodsToAdd, strings.Join(currentMethod, "\n"))
			}
		}
	}

	// For simplicity, we assume that the existing file has a header (package and import sections)
	// and then the method definitions. Here we simply append any new methods to the end.
	merged := existing + "\n\n" + strings.Join(methodsToAdd, "\n\n")
	return merged
}
