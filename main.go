package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/worker"
)

type helloRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	app := sdk.FunctionApp()

	app.HTTP(
		"httpget",
		httpGet,
		sdk.WithMethods("GET"),
		sdk.WithAuth("function"),
	)

	app.HTTP(
		"httppost",
		httpPost,
		sdk.WithMethods("POST"),
		sdk.WithAuth("function"),
	)

	worker.Start(app)
}

func httpGet(w http.ResponseWriter, r *http.Request) {
	ic, _ := sdk.FromContext(r.Context())

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "World"
	}

	attrs := append([]any{
		"method", r.Method,
		"path", r.URL.Path,
		"name", name,
	}, invocationAttrs(ic)...)

	slog.InfoContext(r.Context(), "handling GET request", attrs...)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hello, %s!", name)
}

func httpPost(w http.ResponseWriter, r *http.Request) {
	ic, _ := sdk.FromContext(r.Context())

	payload, err := decodeHelloRequest(r)
	if err != nil {
		attrs := append([]any{
			"method", r.Method,
			"path", r.URL.Path,
			"error", err.Error(),
		}, invocationAttrs(ic)...)

		slog.WarnContext(r.Context(), "invalid POST request", attrs...)
		writeTextResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	attrs := append([]any{
		"method", r.Method,
		"path", r.URL.Path,
		"name", payload.Name,
		"age", payload.Age,
	}, invocationAttrs(ic)...)

	slog.InfoContext(r.Context(), "handling POST request", attrs...)
	writeTextResponse(w, http.StatusOK, fmt.Sprintf("Hello, %s! You are %d years old!", payload.Name, payload.Age))
}

func decodeHelloRequest(r *http.Request) (helloRequest, error) {
	defer r.Body.Close()

	var payload helloRequest

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			return helloRequest{}, errors.New("request body must contain JSON")
		}
		return helloRequest{}, fmt.Errorf("invalid JSON in request body: %w", err)
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return helloRequest{}, errors.New("request body must contain a single JSON object")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return helloRequest{}, errors.New("please provide a non-empty 'name' in the request body")
	}

	if payload.Age <= 0 {
		return helloRequest{}, errors.New("please provide a positive integer 'age' in the request body")
	}

	return payload, nil
}

func invocationAttrs(ic *sdk.InvocationContext) []any {
	if ic == nil {
		return nil
	}

	return []any{
		"invocation_id", ic.InvocationID,
		"function_id", ic.FunctionID,
		"function_name", ic.FunctionName,
		"trigger_type", ic.TriggerType,
		"trace_parent", ic.TraceContext.TraceParent,
	}
}

func writeTextResponse(w http.ResponseWriter, statusCode int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(body))
}
