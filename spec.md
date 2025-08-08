## 📄 Specification Document: `tmpltr`

### 🧩 Overview
`tmpltr` is a command-line tool written in Go that enables users to create, save, and restore directory templates. It supports saving files with hashed identifiers for deduplication, optionally ignoring file contents, and preserving relative paths. Templates can be reused to scaffold new projects or environments.

---

### 🚀 Goals
- Scaffold directory structures with hashed file identifiers
- Save and restore templates with optional file contents
- Provide intuitive CLI commands for managing templates
- Ensure portability and performance using Go

---

### 🛠️ Core Features

#### 1. **Create Template**
```bash
tmpltr make <target_directory> --name="<template_name>" [--ignore-contents]
```
- Recursively scans the target directory
- Saves structure and files to a template store
- Hashes file contents for deduplication
- If `--ignore-contents` is passed, only filenames and structure are saved

#### 2. **Restore Template**
```bash
tmpltr restore --name="<template_name>" --output="<destination_directory>"
```
- Recreates the saved directory structure
- Restores files using hash-based lookup
- If contents were ignored, creates empty files or placeholders

#### 3. **List Templates**
```bash
tmpltr list
```
- Displays all saved templates with metadata

#### 4. **Delete Template**
```bash
tmpltr delete --name="<template_name>"
```
- Removes a saved template from the store

---

### 📦 Template Storage

- Default location: `~/.tmpltr/templates/`
- Each template stored as:
  ```
  ~/.tmpltr/templates/<template_name>/
    ├── manifest.json
    ├── files/
    │   ├── a1b2c3...
    │   ├── d4e5f6...
    │   └── ...
  ```

---

### 🔐 Hashing

- **Purpose:**  
  Hashes are used as unique identifiers for file contents to enable deduplication and efficient lookup. They do *not* replace the actual contents in the stored template.

- **Algorithm:**  
  SHA256 is used to compute a hash of each file’s contents.

- **Behavior:**
  - During template creation:
    - Each file’s contents are hashed.
    - The hash is used as the filename in the internal `files/` store.
    - The actual contents are saved in that file (unless `--ignore-contents` is used).
    - The `manifest.json` maps original relative paths to their content hashes.
  - During restoration:
    - The manifest is read.
    - Files are reconstructed using the hash-to-content mapping.
    - Original filenames and paths are restored.

- **Ignore Mode:**
  - If `--ignore-contents` is used:
    - Hashes may be derived from filenames or generated as UUIDs.
    - Corresponding files in `files/` may be empty or omitted.

---

### 📄 Manifest File

#### `manifest.json` structure:
```json
{
  "name": "my-template",
  "created_at": "2025-08-07T22:58:00Z",
  "files": [
    {
      "original_path": "src/main.go",
      "hash": "a1b2c3...",
      "include_contents": true
    },
    {
      "original_path": "README.md",
      "hash": "d4e5f6...",
      "include_contents": false
    }
  ]
}
```

- `original_path`: Relative path of the file in the source directory
- `hash`: SHA256 hash used for lookup in the `files/` store
- `include_contents`: Boolean flag indicating whether the file contents were saved

---

### 🧪 Optional Enhancements
- `--dry-run` flag to preview actions
- `--verbose` flag for detailed logging
- Support for `.tmpltrignore` file to exclude paths
- Compression of templates for portability

---

### 🧰 Tech Stack
- **Language:** Go
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **Hashing:** `crypto/sha256`
- **File I/O:** `os`, `io`, `path/filepath`
- **JSON Handling:** `encoding/json`

