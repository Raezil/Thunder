package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	// Define command line flags
	var (
		module      = flag.String("module", "", "Go module path for go_package")
		pkg         = flag.String("package", "", "Proto package name")
		serviceName = flag.String("service", "", "gRPC service name")
		entity      = flag.String("entity", "", "Entity name (singular)")
		fields      = flag.String("fields", "", "Entity fields in format 'name:type,email:string,age:int32'")
		outDir      = flag.String("out", ".", "Output directory")
		help        = flag.Bool("help", false, "Show help message")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "thunder-scaffold - Generate a .proto scaffold for full CRUD with gRPC, REST and GraphQL services\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -service UserService -entity User -package users -module github.com/myorg/myapp -fields 'name:string,email:string,age:int32' -out ./proto\n", os.Args[0])
	}

	// Parse command line flags
	flag.Parse()

	// Show help if requested
	if *help {
		flag.Usage()
		return
	}

	// Validate required flags
	if *serviceName == "" || *entity == "" {
		fmt.Fprintf(os.Stderr, "Error: flags -service and -entity are required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse fields if provided
	var entityFields []Field
	if *fields != "" {
		var err error
		entityFields, err = parseFields(*fields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing fields: %v\n\n", err)
			flag.Usage()
			os.Exit(1)
		}
	}

	// Prepare template data
	data := struct {
		Module       string
		Package      string
		ServiceName  string
		Entity       string
		EntityLower  string
		EntityPlural string
		Fields       []Field
		HasFields    bool
	}{
		Module:       *module,
		Package:      *pkg,
		ServiceName:  *serviceName,
		Entity:       *entity,
		EntityLower:  strings.ToLower(*entity),
		EntityPlural: strings.ToLower(*entity) + "s",
		Fields:       entityFields,
		HasFields:    len(entityFields) > 0,
	}

	// Parse template
	tmpl, err := template.New("proto").Parse(protoTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create output file
	filePath := filepath.Join(*outDir, data.EntityLower+".proto")
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Execute template
	if err := tmpl.Execute(f, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated %s successfully!\n", filePath)
}

// Field represents a proto field
type Field struct {
	Name string
	Type string
	Tag  int
}

// parseFields parses field definitions like "name:string,email:string,age:int32"
func parseFields(fieldsStr string) ([]Field, error) {
	if fieldsStr == "" {
		return nil, nil
	}

	fieldPairs := strings.Split(fieldsStr, ",")
	fields := make([]Field, 0, len(fieldPairs))

	for i, pair := range fieldPairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field format '%s', expected 'name:type'", pair)
		}

		name := strings.TrimSpace(parts[0])
		fieldType := strings.TrimSpace(parts[1])

		if name == "" || fieldType == "" {
			return nil, fmt.Errorf("field name and type cannot be empty in '%s'", pair)
		}

		fields = append(fields, Field{
			Name: name,
			Type: fieldType,
			Tag:  i + 2, // Start from 2 since id takes tag 1
		})
	}

	return fields, nil
}

const protoTemplate = `syntax = "proto3";
package {{.Package}};
option go_package = "{{.Module}}/{{.Package}};{{.Package}}";

import "google/api/annotations.proto";
import "graphql.proto";

service {{.ServiceName}} {
  option (graphql.service) = {
    host: "localhost:50051"
    insecure: true
  };

  // Read single
  rpc Get{{.Entity}} (Get{{.Entity}}Request) returns (Get{{.Entity}}Response) {
    option (google.api.http) = {
      get: "/v1/{{.Package}}/{{.EntityLower}}/{id}"
    };
    option (graphql.schema) = {
      type: QUERY
      name: "get{{.Entity}}"
    };
  }

  // Read list
  rpc List{{.Entity}}s (List{{.Entity}}Request) returns (List{{.Entity}}Response) {
    option (google.api.http) = {
      get: "/v1/{{.EntityLower}}s/list"
    };
    option (graphql.schema) = {
      type: QUERY
      name: "list{{.Entity}}s"
    };
  }

  // Create
  rpc Create{{.Entity}} (Create{{.Entity}}Request) returns (Create{{.Entity}}Response) {
    option (google.api.http) = {
      post: "/v1/{{.Package}}"
      body: "*"
    };
    option (graphql.schema) = {
      type: MUTATION
      name: "create{{.Entity}}"
    };
  }

  // Update
  rpc Update{{.Entity}} (Update{{.Entity}}Request) returns (Update{{.Entity}}Response) {
    option (google.api.http) = {
      put: "/v1/{{.Package}}/{{.EntityLower}}/{id}"
      body: "*"
    };
    option (graphql.schema) = {
      type: MUTATION
      name: "update{{.Entity}}"
    };
  }

  // Delete
  rpc Delete{{.Entity}} (Delete{{.Entity}}Request) returns (Delete{{.Entity}}Response) {
    option (google.api.http) = {
      delete: "/v1/{{.Package}}/{{.EntityLower}}/{id}"
    };
    option (graphql.schema) = {
      type: MUTATION
      name: "delete{{.Entity}}"
    };
  }
}

{{if .HasFields}}
// Entity definition
message {{.Entity}} {
  string id = 1;
{{- range .Fields}}
  {{.Type}} {{.Name}} = {{.Tag}};
{{- end}}
}
{{end}}

message Get{{.Entity}}Request {
  string id = 1 [(graphql.field) = {required: true}];
}

message Get{{.Entity}}Response {
  {{.Entity}} {{.EntityLower}} = 1;
}

message List{{.Entity}}Request {}

message List{{.Entity}}Response {
  repeated {{.Entity}} {{.EntityPlural}} = 1;
}

message Create{{.Entity}}Request {
  {{.Entity}} {{.EntityLower}} = 1 [(graphql.field) = {required: true}];
}

message Create{{.Entity}}Response {
  {{.Entity}} {{.EntityLower}} = 1;
}

message Update{{.Entity}}Request {
  {{.Entity}} {{.EntityLower}} = 1 [(graphql.field) = {required: true}];
}

message Update{{.Entity}}Response {
  {{.Entity}} {{.EntityLower}} = 1;
}

message Delete{{.Entity}}Request {
  string id = 1 [(graphql.field) = {required: true}];
}

message Delete{{.Entity}}Response {
  string id = 1;
}
`
