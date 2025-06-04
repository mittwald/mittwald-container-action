# 🐳 mittwald-container-action

This GitHub Action updates a [mittwald](https://mittwald.de) container stack via the official [Container API](https://developer.mittwald.de/docs/v2/reference/container), using the [Go SDK](https://github.com/mittwald/api-client-go).

It supports flexible configuration using YAML files or inline YAML strings and includes support for templating using environment variables.

> ⚠️ Disclaimer: This action is still under active development. Bugs or breaking changes may occur — please test carefully before production use.

---

## 🔧 Features

- Authentication using Mittwald API tokens
- Updates full container stacks using `UpdateStack`
- Supports stack configuration via:
  - single `stack.yaml` (services + volumes)
  - separate `services.yaml` and `volumes.yaml`
  - inline YAML via action `with:` parameters
- Supports environment variable templating in YAML (`{{ .Env.MY_VAR }}`)

---

> ⚠️ **Important Note about the Studio API**
>
> This action uses the `UpdateStack` endpoint of the [mittwald Studio Container API](https://developer.mittwald.de/docs/v2/reference/container).
>
> This means **every service in your stack is declared exactly as described in your YAML file** – including ports, volumes, and environment variables.
>
> 🧨 **Any manual changes made in the mStudio UI that are not reflected in your YAML configuration will be overwritten!**
>
> If you need to preserve manual adjustments, make sure to incorporate them into your version-controlled YAML files before deploying.

---

## 🚀 Usage

### ✅ Minimal example with `stack.yaml`

```yaml
name: Deploy Stack

on:
  push:
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Mittwald
        uses: mittwald/mittwald-container-action@main
        env:
          MYSQL_ROOT_PASSWORD: "${{ secrets.ROOT_PASS }}"
          MYSQL_DATABASE: "testdb"
          MYSQL_USER: "user"
          MYSQL_PASSWORD: "${{ secrets.USER_PASS }}"
        with:
          api_token: ${{ secrets.MITTWALD_API_TOKEN }}
          stack_id: "your-stack-id"
          stack_file: "${{ github.workspace }}/configs/stack.yaml"
```

---

### 🧪 Example: Full stack file (`stack.yaml`)

```yaml
services:
  mydb:
    image: "mysql:8.0"
    description: "MySQL"
    ports:
      - "3306/tcp"
    envs:
      MYSQL_ROOT_PASSWORD: {{ .Env.MYSQL_ROOT_PASSWORD }}
      MYSQL_DATABASE: {{ .Env.MYSQL_DATABASE }}
      MYSQL_USER: {{ .Env.MYSQL_USER }}
      MYSQL_PASSWORD: {{ .Env.MYSQL_PASSWORD }}
    volumes:
      - "mysql:/var/lib/mysql"

volumes:
  mysql:
    name: mydb-volume
```

ℹ️ Templating: The placeholders like `{{ .Env.MYSQL_ROOT_PASSWORD }}` will be replaced using the environment variables at runtime.

You must provide these variables in the workflow's `env` block if they are used in your stack file:

```yaml
env:
  MYSQL_ROOT_PASSWORD: "rootpass"
  MYSQL_DATABASE: "testdb"
  MYSQL_USER: "user"
  MYSQL_PASSWORD: "secret"
```

---

### 🧪 Example: Separate `services.yaml` and `volumes.yaml`

```yaml
with:
  api_token: ${{ secrets.MITTWALD_API_TOKEN }}
  stack_id: "your-stack-id"
  services_file: "${{ github.workspace }}/configs/services.yaml"
  volumes_file: "${{ github.workspace }}/configs/volumes.yaml"
```

---

### 🧪 Example: Inline YAML in workflow

```yaml
with:
  api_token: ${{ secrets.MITTWALD_API_TOKEN }}
  stack_id: "your-stack-id"
  services_yaml: |
    app:
      image: "nginx"
      description: "Nginx"
      ports:
        - "80/tcp"
```

---

## ⚙️ Inputs

| Name                        | Required | Description                                  |
|-----------------------------|----------|----------------------------------------------|
| `api_token`                 | ✅       | Mittwald API token                           |
| `stack_id`                  | ✅       | Stack UUID to update                         |
| `stack_yaml` / `stack_file`        | 🔄       | Full stack (services + volumes) YAML         |
| `services_yaml` / `services_file`  | 🔄       | Services-only YAML                           |
| `volumes_yaml` / `volumes_file`    | 🔄       | Volumes-only YAML                            |

---

## 🧩 Environment Variable Parsing

This action supports dynamic configuration using environment variables inside YAML files.

Example usage in YAML:

```yaml
envs:
  MONGODB_PASSWORD: {{ .Env.MONGODB_PASSWORD }}
```

To make this work, you must define the variables in your workflow under the `env:` section:

```yaml
env:
  MONGODB_PASSWORD: ${{ secrets.DB_PASSWORD }}
```

---

## 🧪 Full Example with Secret Templating

```yaml
name: Deploy Mittwald Stack

on:
  push:
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest

    env:
      MONGODB_ROOT_PASSWORD: ${{ secrets.MONGO_ROOT_PW }}
      ME_CONFIG_BASICAUTH_PASSWORD: ${{ secrets.MONGOEXPRESS_PW }}

    steps:
      - name: Checkout project
        uses: actions/checkout@v4

      - name: Deploy stack to Mittwald
        uses: mittwald/mittwald-container-action@main
        with:
          api_token: ${{ secrets.MITTWALD_API_TOKEN }}
          stack_id: "your-stack-id"
          stack_file: "${{ github.workspace }}/configs/mstudio/container_stack.yaml"
```

---

## 📁 Examples

Several full example files are provided under the `examples/` directory in this repository.

---

## 🛠 Development

This action is written in Go and uses the official [mittwald/api-client-go](https://github.com/mittwald/api-client-go) SDK (v2). See `main.go` for details.

---

## 🤝 Contributing

Feel free to open issues or PRs — improvements and use-cases are welcome!
