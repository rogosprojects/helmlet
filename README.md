# Helmlet

## 📝 What is Helmlet?

Helmlet is a lightweight, fast, and flexible templating engine for YAML files, designed as a simpler alternative to Helm for Kubernetes manifests. While Helm offers a complete package management solution for Kubernetes, Helmlet focuses solely on the templating aspect with minimal dependencies.

## ❓ Why use Helmlet?

- **Simplicity**: Focused only on templating without the complexity of a full package manager
- **Speed**: Minimal dependencies and fast execution for quick rendering
- **Flexibility**: Support for multiple templates and value files with intuitive overriding
- **Familiarity**: Uses the same templating syntax as Helm for easy adoption
- **Portability**: Single binary with no external dependencies

## Core Functionality

Helmlet takes values from YAML files and/or command-line parameters, processes one or more template files, and generates rendered YAML output suitable for direct application to Kubernetes or other YAML-based systems.

## 🚀 Quick Start

Get started with Helmlet in minutes:

1. **Download the binary or build from source**
   ```bash
   # Build from source
   go build -ldflags "-X main.Version=$(git describe --tags --always)" -o helmlet
   ```

2. **Create a simple template (template.yaml)**
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: {{ .Values.name }}
   spec:
     replicas: {{ .Values.replicas }}
     selector:
       matchLabels:
         app: {{ .Values.name }}
     template:
       metadata:
         labels:
           app: {{ .Values.name }}
       spec:
         containers:
         - name: {{ .Values.name }}
           image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
           ports:
           - containerPort: {{ .Values.containerPort }}
   ```

3. **Create a values file (values.yaml)**
   ```yaml
   name: myapp
   replicas: 3
   image:
     repository: nginx
     tag: latest
   containerPort: 80
   ```

4. **Render the template**
   ```bash
   ./helmlet --template template.yaml --value values.yaml
   ```

5. **Apply directly to Kubernetes (optional)**
   ```bash
   ./helmlet --template template.yaml --value values.yaml | kubectl apply -f -
   ```

## Command Line Options

| Flag | Description | Example |
|------|-------------|---------|
| `--template` | Template file(s) to process (can be specified multiple times) | `--template deployment.yaml` |
| `--template-dir` | Directory containing template files | `--template-dir k8s-templates/` |
| `--value` | Values file(s) to load (can be specified multiple times) | `--value values.yaml` |
| `--set` | Set values on the command line | `--set image.tag=v1.2.3,replicas=3` |
| `--output` | Write output to file (for single templates) | `--output rendered.yaml` |
| `--output-dir` | Directory to write output files (for multiple templates) | `--output-dir rendered/` |
| `--delimiter` | Custom template delimiters | `--delimiter '<%,%>'` |
| `--strict` | If true, fail when template uses missing keys | `--strict` |

## 📚 Examples

### Basic Usage

```bash
# Simple template rendering with command line values
./helmlet --template template.yaml --set key1=value1,key2=value2

# Render with values from file
./helmlet --template template.yaml --value values.yaml

# Output to file instead of stdout
./helmlet --template template.yaml --value values.yaml --output rendered.yaml
```

### Advanced Usage

```bash
# Layer multiple values files (values from later files override earlier ones)
./helmlet --template template.yaml --value common.yaml --value dev.yaml --value app.yaml

# Process an entire directory of templates
./helmlet --template-dir templates/ --value values.yaml --output-dir rendered/

# Process multiple specific templates
./helmlet --template template1.yaml --template template2.yaml --value values.yaml --output-dir rendered/

# Use quoted values to include commas in values
./helmlet --template template.yaml --set 'key1="value1,has,commas"'

# Use custom delimiters to avoid conflicts
./helmlet --template template.yaml --value values.yaml --delimiter '<%,%>'

# Enable strict mode to fail on missing values
./helmlet --template template.yaml --value values.yaml --strict
```

### 🛠️ Integration with Kubernetes

```bash
# Render template and pipe directly to kubectl
./helmlet --template deployment.yaml --value values.yaml | kubectl apply -f -

# Process templates and save for later use
./helmlet --template-dir k8s-templates/ --value values.yaml --output-dir k8s-manifests/
kubectl apply -f k8s-manifests/
```

## 🌟 Features

### Core Features
- **Templating Engine**: Go templates with custom functions similar to Helm
- **Multiple Template Processing**: Process individual files or entire directories
- **Value Management**:
  - Layer multiple value files (later files override earlier ones)
  - Set individual values via command line
  - Support for nested values with dot notation (e.g., `app.config.port=8080`)
- **Output Options**:
  - Direct to stdout for piping to kubectl or other tools
  - Single output file for individual templates
  - Output directory for batch processing multiple templates

### Template Functions Supported
- **default**: Provide default values when variables are undefined
- **quote**: Properly quote string values
- **indent**: Control indentation for multiline strings
- **tpl**: Dynamically process strings as templates
- **FilesGet**: Include external files within templates

### Special Features
- **Strict Mode**: Optional enforcement of required template variables
- **Custom Delimiters**: Change template delimiters to avoid conflicts with target formats
- **UTF-8 Conversion**: Automatically detects and converts non-UTF8 files

## 🤝 Contributing

We welcome contributions! Please feel free to submit a pull request (PR) or open an issue on GitHub.



