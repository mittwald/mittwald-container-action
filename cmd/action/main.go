package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/pkg/errors"

	"github.com/mittwald/api-client-go/mittwaldv2"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/containerclientv2"
	"gopkg.in/yaml.v3"
)

func main() {
	apiToken := mustEnv("INPUT_API_TOKEN")
	stackID := mustEnv("INPUT_STACK_ID")
	ctx := context.Background()

	client, createClientErr := mittwaldv2.New(ctx, mittwaldv2.WithAccessToken(apiToken))
	if createClientErr != nil {
		slog.With(slog.Any("error", createClientErr)).Error("error creating mittwaldv2 client")

		os.Exit(1)
	}

	stackData, loadStackDataErr := loadStackData()
	if loadStackDataErr != nil {
		slog.With(slog.Any("error", loadStackDataErr)).Error("❌ failed to load stack data")

		os.Exit(1)
	}
	if validateErr := stackData.Validate(); validateErr != nil {
		slog.With(slog.Any("error", validateErr)).Error("❌ invalid stack data")
	}

	req := containerclientv2.DeclareStackRequest{
		Body:    *stackData,
		StackID: stackID,
	}

	stackResponse, httpResponse, declareStackErr := client.Container().DeclareStack(ctx, req)
	if declareStackErr != nil {
		slog.With(slog.Any("error", declareStackErr)).Error("❌ failure while declaring stack")

		plainHttpResponse, plainHttpResponseErr := io.ReadAll(httpResponse.Body)
		if plainHttpResponseErr == nil {
			slog.With(slog.Any("response", string(plainHttpResponse))).Error("🔎 http-response")
		}

		os.Exit(1)
	}

	slog.With(slog.Any("stackResponse", stackResponse)).Info("✅ Stack updated successfully")
}

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

func parseStackObject(raw map[string]interface{}) (*containerclientv2.DeclareStackRequestBody, error) {
	data, marshalErr := yaml.Marshal(raw)
	if marshalErr != nil {
		return nil, errors.Wrap(marshalErr, "failed to marshal inputs")
	}

	var stack containerclientv2.DeclareStackRequestBody
	if unmarshalErr := yaml.Unmarshal(data, &stack); unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "failure while unmarshalling inputs to declareStackRequestBody")
	}

	return &stack, nil
}

func loadYamlOptional(name string) (map[string]interface{}, error) {
	file := os.Getenv("INPUT_" + name + "_FILE")
	raw := os.Getenv("INPUT_" + name + "_YAML")

	var bytes []byte

	if file != "" {
		var readFileErr error
		bytes, readFileErr = os.ReadFile(file)
		if readFileErr != nil {
			return nil, errors.Wrap(readFileErr, "failure while reading file "+file)
		}
	} else if raw != "" {
		bytes = []byte(raw)
	} else {
		return nil, nil
	}

	var parsed map[string]interface{}
	if unmarshalErr := yaml.Unmarshal(bytes, &parsed); unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "failure while unmarshalling data")
	}

	return parsed, nil
}

func loadYamlRequired(name string) (map[string]interface{}, error) {
	data, readDataErr := loadYamlOptional(name)
	if data == nil && readDataErr == nil {
		return nil, errors.WithStack(errors.New("unable to locate inputs: neither _FILE nor _YAML specified"))
	}

	return data, readDataErr
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("❌ Missing required environment variable " + key)
	}
	return val
}
