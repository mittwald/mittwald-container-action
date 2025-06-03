# 🐳 mittwald-container-action

This GitHub Action updates a container stack on [mittwald. mStudio](https://mittwald.de) using the official [mittwald. Container API](https://developer.mittwald.de/docs/v2/reference/container) via their [Go SDK](https://github.com/mittwald/api-client-go).

It supports flexible configuration via inline YAML or YAML files, and is compatible with complex, multi-service stacks.

---

> ⚠️ **Disclaimer**  
> This action is currently under active development.  
> Features may change and bugs may occur.  
> We recommend testing in a staging environment before using in production.

---

## 🔧 What It Does

- Uses the [`mittwald/api-client-go`](https://github.com/mittwald/api-client-go) SDK (v2)
- Authenticates via API token
- Calls the `DeclareStack` endpoint to update a container stack
- Accepts stack configuration via:
  - a combined `stack.yaml` (with `services` + `volumes`)
  - separate `services` and `volumes` YAML files or strings

---

## 🚀 How to Use

### 🧪 Minimal Example

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Update mittwald. container stack
        uses: mittwald/mittwald-container-action@main
        with:
          api_token: ${{ secrets.MITTWALD_API_TOKEN }}
          stack_id: "your-stack-id"
          stack_file: "./examples/stack.yaml"
