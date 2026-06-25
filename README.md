<!--
---
name: Azure Functions Go HTTP Trigger using Azure Developer CLI
description: This repository contains an Azure Functions HTTP trigger quickstart written in Go and deployed to Azure Functions Flex Consumption using the Azure Developer CLI (azd). The sample uses managed identity and a virtual network to make sure deployment is secure by default. You can opt out of a VNet being used in the sample by setting VNET_ENABLED to false in the parameters.
page_type: sample
languages:
- azdeveloper
- go
- bicep
products:
- azure
- azure-functions
- entra-id
urlFragment: functions-quickstart-go-azd
---
-->

# Azure Functions Go HTTP Trigger using Azure Developer CLI

This template repository contains an HTTP trigger reference sample for Azure Functions written in Go and deployed to Azure using the Azure Developer CLI (`azd`). The sample uses managed identity and a virtual network to make sure deployment is secure by default.

This source code supports the article [Quickstart: Create and deploy functions to Azure Functions using the Azure Developer CLI](https://learn.microsoft.com/azure/azure-functions/create-first-function-azure-developer-cli?pivots=programming-language-go).

## Prerequisites

+ [Go 1.24 or later](https://go.dev/dl/)
+ [Azurite](https://learn.microsoft.com/azure/storage/common/storage-use-azurite) (for local storage emulation)
+ [Azure Functions Core Tools](https://learn.microsoft.com/azure/azure-functions/functions-run-local?pivots=programming-language-go#install-the-azure-functions-core-tools) 4.12.0 or later
+ [Azure Developer CLI (AZD)](https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd)
+ To use Visual Studio Code to run and debug locally:
  + [Visual Studio Code](https://code.visualstudio.com/)
  + [Azure Functions extension](https://marketplace.visualstudio.com/items?itemName=ms-azuretools.vscode-azurefunctions)
  + [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)

## Initialize the local project

You can initialize a project from this `azd` template in one of these ways:

+ Use this `azd init` command from an empty local (root) folder:

    ```shell
    azd init --template functions-quickstart-go-azd
    ```

    Supply an environment name, such as `flexquickstart` when prompted. In `azd`, the environment is used to maintain a unique deployment context for your app.

+ Clone the GitHub template repository locally using the `git clone` command:

    ```shell
    git clone https://github.com/Azure-Samples/functions-quickstart-go-azd.git
    cd functions-quickstart-go-azd
    ```

    You can also clone the repository from your own fork in GitHub.

## Local settings

The `local.settings.json` file is included in this template with default values for local development. This file is excluded from deployment by `.funcignore`. The local project uses `FUNCTIONS_WORKER_RUNTIME` set to `native`, which is the required worker runtime for the Go worker.

## Run your app from the terminal

1. Start Azurite for local storage emulation. In a separate terminal, run:

    ```shell
    azurite
    ```

1. In your project terminal, restore dependencies and start the Functions host:

    ```shell
    go mod tidy
    func start
    ```

    `func start` automatically builds the Go project before launching the local Azure Functions host.

1. From a new terminal, call the HTTP GET endpoint:

    ```shell
    curl -i "http://localhost:7071/api/httpget?name=Azure"
    ```

1. Test the HTTP POST trigger with a payload using your favorite secure HTTP test tool. This example uses the `curl` tool with payload data from the [`testdata.json`](./testdata.json) project file:

    ```shell
    curl -i http://localhost:7071/api/httppost -H "Content-Type: application/json" -d @testdata.json
    ```

1. When you're done, press Ctrl+C in the terminal window to stop the `func.exe` host process.

## Run your app using Visual Studio Code

1. Open the root folder in a new terminal.
1. Run the `code .` command to open the project in Visual Studio Code.
1. Press **Run/Debug (F5)** to start the Functions host and attach the Go debugger.
1. Send GET and POST requests to the `httpget` and `httppost` endpoints respectively using your HTTP test tool (or browser for `httpget`). If you have the [RestClient](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension installed, you can execute requests directly from the [`test.http`](test.http) project file.

## Source Code

The source code for both functions is in the [`main.go`](./main.go) code file. Azure Functions uses the native Go worker SDK, so no `function.json` files are required because the worker indexes the app directly from code.

This code shows an HTTP GET triggered function:

```go
func httpGet(w http.ResponseWriter, r *http.Request) {
    ic, _ := sdk.FromContext(r.Context())

    name := strings.TrimSpace(r.URL.Query().Get("name"))
    if name == "" {
        name = "World"
    }

    slog.InfoContext(r.Context(), "handling GET request",
        "name", name,
        "invocation_id", ic.InvocationID,
    )

    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Hello, %s!", name)
}
```

This code shows an HTTP POST triggered function:

```go
func httpPost(w http.ResponseWriter, r *http.Request) {
    payload, err := decodeHelloRequest(r)
    if err != nil {
        writeTextResponse(w, http.StatusBadRequest, err.Error())
        return
    }

    slog.InfoContext(r.Context(), "handling POST request",
        "name", payload.Name,
        "age", payload.Age,
    )

    writeTextResponse(w, http.StatusOK, fmt.Sprintf("Hello, %s! You are %d years old!", payload.Name, payload.Age))
}
```

## Deploy to Azure

Run this command to provision the function app, with any required Azure resources, and deploy your code:

```shell
azd up
```

By default, this sample deploys with a virtual network (VNet) for enhanced security, ensuring that the function app and related resources are isolated within a private network. 
The `VNET_ENABLED` parameter controls whether a VNet is used during deployment:
- When `VNET_ENABLED` is `true` (default), the function app is deployed with a VNet for secure communication and resource isolation.
- When `VNET_ENABLED` is `false`, the function app is deployed without a VNet, allowing public access to resources.

To disable the VNet for this sample, set `VNET_ENABLED` to `false` before running `azd up`:

```bash
azd env set VNET_ENABLED false
azd up
```

You're prompted to supply these required deployment parameters:

| Parameter | Description |
| ---- | ---- |
| _Environment name_ | An environment that's used to maintain a unique deployment context for your app. You aren't prompted when you created the local project using `azd init`.|
| _Azure subscription_ | Subscription in which your resources are created.|
| _Azure location_ | Azure region in which to create the resource group that contains the new Azure resources. Only regions that currently support the Flex Consumption plan are shown.|

To learn how to obtain your new function endpoints in Azure along with the required function keys, see [Invoke the function on Azure](https://learn.microsoft.com/azure/azure-functions/create-first-function-azure-developer-cli?pivots=programming-language-go#invoke-the-function-on-azure) in the companion article [Quickstart: Create and deploy functions to Azure Functions using the Azure Developer CLI](https://learn.microsoft.com/azure/azure-functions/create-first-function-azure-developer-cli?pivots=programming-language-go).

## Redeploy your code

You can run the `azd up` command as many times as you need to both provision your Azure resources and deploy code updates to your function app.

>[!NOTE]
>Deployed code files are always overwritten by the latest deployment package.

## Clean up resources

When you're done working with your function app and related resources, you can use this command to delete the function app and its related resources from Azure and avoid incurring any further costs:

```shell
azd down
```
