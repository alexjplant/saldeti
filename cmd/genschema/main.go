// Package main generates a JSON Schema from SeedConfig and validates the example seed file.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	invopopjsonschema "github.com/invopop/jsonschema"
	"github.com/saldeti/saldeti/internal/seed"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func main() {
	// 1. Generate JSON Schema from SeedConfig
	reflector := &invopopjsonschema.Reflector{}
	schema := reflector.Reflect(&seed.SeedConfig{})

	// 2. Ensure schema/ directory exists
	if err := os.MkdirAll("schema", 0o755); err != nil {
		fatalf("creating schema directory: %v", err)
	}

	// 3. Marshal schema to indented JSON
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fatalf("marshalling schema: %v", err)
	}

	// 4. Write schema to schema/seed.schema.json (trailing newline)
	schemaPath := "schema/seed.schema.json"
	if err := os.WriteFile(schemaPath, append(schemaBytes, '\n'), 0o644); err != nil {
		fatalf("writing schema file: %v", err)
	}
	fmt.Printf("Generated schema: %s\n", schemaPath)

	// 5. Validate examples/seed.json against the generated schema
	absSchemaPath, err := filepath.Abs(schemaPath)
	if err != nil {
		fatalf("resolving schema path: %v", err)
	}

	sch, err := jsonschema.Compile("file://" + absSchemaPath)
	if err != nil {
		fatalf("compiling schema: %v", err)
	}

	exampleBytes, err := os.ReadFile("examples/seed.json")
	if err != nil {
		fatalf("reading examples/seed.json: %v", err)
	}

	var exampleData interface{}
	if err := json.Unmarshal(exampleBytes, &exampleData); err != nil {
		fatalf("parsing examples/seed.json: %v", err)
	}

	if err := sch.Validate(exampleData); err != nil {
		fatalf("examples/seed.json does not validate against schema:\n%v", err)
	}

	fmt.Println("examples/seed.json validates against the schema ✓")
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
