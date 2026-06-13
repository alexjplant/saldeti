// Package main generates JSON Schemas from seed config types and validates example seed files.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	invopopjsonschema "github.com/invopop/jsonschema"
	"github.com/saldeti/saldeti/internal/entra/seed"
	gseed "github.com/saldeti/saldeti/internal/google/seed"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func main() {
	// 1. Generate Entra ID JSON Schema from SeedConfig
	reflector := &invopopjsonschema.Reflector{}
	entraSchema := reflector.Reflect(&seed.SeedConfig{})

	if err := os.MkdirAll("schema", 0o755); err != nil {
		fatalf("creating schema directory: %v", err)
	}

	entraSchemaBytes, err := json.MarshalIndent(entraSchema, "", "  ")
	if err != nil {
		fatalf("marshalling Entra schema: %v", err)
	}

	entraSchemaPath := "schema/seed.schema.json"
	if err := os.WriteFile(entraSchemaPath, append(entraSchemaBytes, '\n'), 0o644); err != nil {
		fatalf("writing Entra schema file: %v", err)
	}
	fmt.Printf("Generated schema: %s\n", entraSchemaPath)

	// 2. Generate Google Workspace JSON Schema from GoogleSeedConfig
	googleSchema := reflector.Reflect(&gseed.GoogleSeedConfig{})

	googleSchemaBytes, err := json.MarshalIndent(googleSchema, "", "  ")
	if err != nil {
		fatalf("marshalling Google schema: %v", err)
	}

	googleSchemaPath := "schema/google-seed.schema.json"
	if err := os.WriteFile(googleSchemaPath, append(googleSchemaBytes, '\n'), 0o644); err != nil {
		fatalf("writing Google schema file: %v", err)
	}
	fmt.Printf("Generated schema: %s\n", googleSchemaPath)

	// 3. Validate examples/seed.json against the Entra schema
	validateExample(entraSchemaPath, "examples/seed.json", "Entra")

	// 4. Validate examples/google-seed.json against the Google schema
	validateExample(googleSchemaPath, "examples/google-seed.json", "Google")
}

func validateExample(schemaPath, examplePath, label string) {
	absSchemaPath, err := filepath.Abs(schemaPath)
	if err != nil {
		fatalf("resolving schema path: %v", err)
	}

	sch, err := jsonschema.Compile("file://" + absSchemaPath)
	if err != nil {
		fatalf("compiling %s schema: %v", label, err)
	}

	exampleBytes, err := os.ReadFile(examplePath)
	if err != nil {
		fatalf("reading %s: %v", examplePath, err)
	}

	var exampleData interface{}
	if err := json.Unmarshal(exampleBytes, &exampleData); err != nil {
		fatalf("parsing %s: %v", examplePath, err)
	}

	if err := sch.Validate(exampleData); err != nil {
		fatalf("%s does not validate against schema:\n%v", examplePath, err)
	}

	fmt.Printf("%s validates against the %s schema\n", examplePath, label)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
