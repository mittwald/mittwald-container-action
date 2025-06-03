package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/mittwald/api-client-go/mittwaldv2"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/containerclientv2"
	"gopkg.in/yaml.v3"
)

func main() {
	// Required: Mittwald API token and stack ID must be provided via GitHub Action inputs
	apiToken := mustEnv("INPUT_API_TOKEN")
	stackID := mustEnv("INPUT_STACK_ID")
	ctx := context.Background()

	// Initialize mittwald API client with access token
	client, createClientErr := mittwaldv2.New(ctx, mittwaldv2.WithAccessToken(apiToken))
	if createClientErr != nil {
		slog.Error("error creating mittwaldv2 client")
		panic(createClientErr)
	}

	// Load the stack configuration (from file or inline YAML)
	stackData, loadStackDataErr := loadStackData()
	if loadStackDataErr != nil {
		slog.Error("❌ failed to load stack data")
		panic(loadStackDataErr)
	}

	// Run type-level validation (SDK-based) on parsed stack config
	if validateErr := stackData.Validate(); validateErr != nil {
		slog.Error("❌ invalid stack data")
		panic(validateErr)
	}

	// Construct the API request to declare the stack
	req := containerclientv2.DeclareStackRequest{
		Body:    *stackData,
		StackID: stackID,
	}

	// Call the API — this overrides the stack with the full state from YAML
	_, httpResponse, declareStackErr := client.Container().DeclareStack(ctx, req)
	if declareStackErr != nil {
		slog.With(slog.Any("error", declareStackErr)).Error("❌ failure while declaring stack")

		// If available, dump the raw HTTP body for diagnostics
		plainHttpResponse, plainHttpResponseErr := io.ReadAll(httpResponse.Body)
		if plainHttpResponseErr == nil {
			slog.With(slog.Any("response", string(plainHttpResponse))).Error("🔎 http-response")
		}

		panic(declareStackErr)
	}

	slog.Info("✅ Stack updated successfully")
}

// loadStackData determines whether a full stack definition or a services/volumes split was provided.
// It then parses the input into a struct that matches the API's expected payload.
func loadStackData() (*containerclientv2.DeclareStackRequestBody, error) {
	stack, loadStackErr := loadYamlOptional("STACK")
	if loadStackErr != nil {
		return nil, loadStackErr
	}
	if stack != nil {
		return parseStackObject(stack)
	}

	services, loadServicesErr := loadYamlRequired("SERVICES")
	if loadServicesErr != nil {
		return nil, loadServicesErr
	}
	volumes, loadVolumesErr := loadYamlOptional("VOLUMES")
	if loadVolumesErr != nil {
		return nil, loadVolumesErr
	}

	return parseStackObject(
		map[string]interface{}{
			"services": services,
			"volumes":  volumes,
		},
	)
}

// parseStackObject marshals and unmarshals YAML-parsed data into a typed SDK struct.
func parseStackObject(raw map[string]interface{}) (*containerclientv2.DeclareStackRequestBody, error) {
	data, marshalErr := json.Marshal(raw)
	if marshalErr != nil {
		return nil, errors.Wrap(marshalErr, "failed to marshal inputs")
	}

	var stack containerclientv2.DeclareStackRequestBody
	if unmarshalErr := json.Unmarshal(data, &stack); unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "failure while unmarshalling inputs to declareStackRequestBody")
	}

	return &stack, nil
}

// loadYamlOptional attempts to load a YAML config from either a _FILE or _YAML input.
// If both are missing, it returns nil.
func loadYamlOptional(name string) (map[string]interface{}, error) {
	file := os.Getenv("INPUT_" + name + "_FILE")
	raw := os.Getenv("INPUT_" + name + "_YAML")

	var rawInput []byte

	if file != "" {
		var readFileErr error
		rawInput, readFileErr = os.ReadFile(file)
		if readFileErr != nil {
			return nil, errors.Wrap(readFileErr, "failure while reading file "+file)
		}
	} else if raw != "" {
		rawInput = []byte(raw)
	} else {
		return nil, nil
	}

	// Parse as Go template to allow environment variable substitution (e.g., {{ .Env.MY_VAR }})
	configTemplate, createTplErr := template.New("").Parse(string(rawInput))
	if createTplErr != nil {
		return nil, errors.Wrap(createTplErr, "failure while creating template from input")
	}

	// Render template using env vars
	templatedInput, parseTemplateErr := renderConfigTemplate(configTemplate)
	if parseTemplateErr != nil {
		return nil, parseTemplateErr
	}

	// Parse templated YAML into map
	var parsed map[string]interface{}
	if unmarshalErr := yaml.Unmarshal(templatedInput.Bytes(), &parsed); unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "failure while unmarshalling data")
	}

	return parsed, nil
}

// loadYamlRequired is like loadYamlOptional, but throws an error if no input is found.
func loadYamlRequired(name string) (map[string]interface{}, error) {
	data, readDataErr := loadYamlOptional(name)
	if data == nil && readDataErr == nil {
		return nil, errors.WithStack(errors.New("unable to locate inputs: neither _FILE nor _YAML specified"))
	}

	return data, readDataErr
}

// mustEnv fetches a required environment variable or exits the process with an error.
func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("❌ Missing required environment variable " + key)
	}
	return val
}

// renderConfigTemplate renders a Go text/template using environment variables as input.
// Template variables can be accessed using {{ .Env.VARNAME }} syntax.
func renderConfigTemplate(configTemplate *template.Template) (*bytes.Buffer, error) {
	type templateData struct {
		Env map[string]string
	}

	data := templateData{
		Env: make(map[string]string),
	}

	// Collect all environment variables as key-value pairs
	for _, env := range os.Environ() {
		e := strings.SplitN(env, "=", 2)
		if len(e) > 1 {
			data.Env[e[0]] = e[1]
		}
	}

	// Render the template into a buffer
	renderedCfg := new(bytes.Buffer)
	templateErr := configTemplate.Execute(renderedCfg, &data)
	if templateErr != nil {
		return nil, errors.Wrap(templateErr, "failure while rendering template")
	}

	return renderedCfg, nil
}
