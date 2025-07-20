# Page Inventory

A Go CLI tool that generates markdown inventory pages for your files based on flexible configuration rules. Perfect for organizing documentation, notes, diagrams, or any collection of files with automated linking and filtering.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Install](#install)
  - [Build from source](#build-from-source)
- [Usage](#usage)
  - [Example](#example)
- [Configuration](#configuration)
  - [Configuration Fields](#configuration-fields)
- [Tag Filtering](#tag-filtering)
  - [1. YAML Frontmatter](#1-yaml-frontmatter)
  - [2. TOML Frontmatter](#2-toml-frontmatter)
  - [3. Inline Tags (with #)](#3-inline-tags-with-)
- [Custom Formatting](#custom-formatting)
  - [Format Examples](#format-examples)
- [Templates](#templates)
  - [Example Template (`templates/diagrams-template.md`)](#example-template-templatesdiagrams-templatemd)
  - [Template with Table Format](#template-with-table-format)
- [Examples](#examples)
  - [Organizing Notes by Topic](#organizing-notes-by-topic)
  - [Code Documentation](#code-documentation)
  - [Multi-Extension Inventory](#multi-extension-inventory)
- [Output Examples](#output-examples)
  - [Default Format Output](#default-format-output)
  - [Table Format Output](#table-format-output)
- [Use Cases](#use-cases)
- [License](#license)
- [Changelog](#changelog)

## Features

- **Multi-task processing**: Define multiple inventory tasks in a single configuration file
- **Flexible filtering**: Filter files by extension and tags (both frontmatter and content)
- **Template support**: Use custom markdown templates for each inventory page
- **Tag detection**: Supports YAML frontmatter (`---`), TOML frontmatter (`+++`), and inline tags (`#tag`)
- **Custom formatting**: Customize how each file entry is displayed
- **Cross-platform**: Works on Windows, macOS, and Linux with proper path handling
- **Modification dates**: Automatically includes last modification dates

## Installation

### Prerequisites

- Go 1.19 or later


### Install

```
go install github.com/tpfeiffer67/MarkdownFileInventory
```

### Build from source

```bash
git clone github.com/tpfeiffer67/MarkdownFileInventory
cd MarkdownFileInventory
go build .
```

## Usage

```bash
./pageinventory <root_folder>
```

The tool looks for a configuration file named `.markdown-file-inventory.yaml` in the specified root folder.

### Watch Mode

To automatically re-run tasks when files change, use the `--watch` flag:

```bash
./pageinventory --watch <root_folder>
```

This is useful for keeping your inventories up-to-date in real-time while you work.

### Example

```bash
./pageinventory ./my-project
```

This will:
1. Read configuration from `./my-project/.markdown-file-inventory.yaml`
2. Process all defined tasks
3. Generate inventory markdown files in the root folder

## Configuration

Create a `.markdown-file-inventory.yaml` file in your project root:

```yaml
tasks:
  - output_file: "diagrams-inventory.md"
    template: "templates/diagrams-template.md"
    folders:
      - "docs/diagrams"
      - "architecture"
    extensions:
      - ".excalidraw.md"
    tags:
      - "diagram"
      - "architecture"
    format: "- [%s](%s) %s\n"

  - output_file: "ai-notes.md"
    template: "templates/ai-template.md"
    folders:
      - "notes"
      - "research"
    extensions:
      - ".md"
    tags:
      - "ia"
      - "ai"
      - "machine-learning"
    format: "📄 **%s** - [View](%s) *(modified %s)*\n"

  - output_file: "all-docs.md"
    folders:
      - "docs"
    extensions:
      - ".md"
    # No tags = include all files
    # No template = no header content
    # No format = use default format
```

### Configuration Fields

| Field | Required | Description |
|-------|----------|-------------|
| `output_file` | ✅ | Name of the generated markdown file |
| `template` | ❌ | Path to template file (relative to root folder) |
| `folders` | ✅ | List of folders to scan (relative to root folder) |
| `extensions` | ✅ | List of file extensions to include |
| `tags` | ❌ | List of tags to filter by (if empty, includes all files) |
| `format` | ❌ | Custom printf format for each file entry |

## Tag Filtering

Files are included if they contain **any** of the specified tags. Tags can be defined in three ways:

### 1. YAML Frontmatter

```markdown
---
title: "My Document"
tags:
  - ia
  - research
  - important
---

# Document content
```

### 2. TOML Frontmatter

```markdown
+++
title = "My Document"
tags = ["ia", "research", "important"]
+++

# Document content
```

### 3. Inline Tags (with #)

```markdown
# My Document

This document covers #ia and #research topics.

#important #note
```

## Custom Formatting

The `format` field uses printf syntax with three parameters:

1. `%s` - File name (without extension)
2. `%s` - File path (URL-encoded)
3. `%s` - Modification date (YYYY-MM-DD)

### Format Examples

```yaml
# Default format
format: "- [%s](%s) %s\n"
# Output: - [filename](path/to/file.md) 2024-12-15

# Table format
format: "| [%s](%s) | %s |\n"
# Output: | [filename](path/to/file.md) | 2024-12-15 |

# Simple list (no date)
format: "- [%s](%s)\n"
# Output: - [filename](path/to/file.md)

# Date first
format: "%s - [%s](%s)\n"
# Output: 2024-12-15 - [filename](path/to/file.md)

# Rich format with emojis
format: "📄 **%s** - [View](%s) *(modified %s)*\n"
# Output: 📄 **filename** - [View](path/to/file.md) *(modified 2024-12-15)*
```

## Templates

Templates are optional markdown files that provide header content for your inventory pages.

### Example Template (`templates/diagrams-template.md`)

```markdown
# System Architecture Diagrams

This page contains all system architecture diagrams created with Excalidraw.

## Available Diagrams

```

### Template with Table Format

For table formats, include the table header in your template:

```markdown
# Documentation Inventory

| Document | Last Modified |
|----------|---------------|
```

Then use format: `"| [%s](%s) | %s |\n"`

## Examples

### Organizing Notes by Topic

```yaml
tasks:
  - output_file: "ai-research.md"
    template: "templates/research-template.md"
    folders:
      - "notes"
      - "research"
      - "papers"
    extensions:
      - ".md"
    tags:
      - "ai"
      - "machine-learning"
      - "deep-learning"

  - output_file: "project-notes.md"
    folders:
      - "notes"
    extensions:
      - ".md"
    tags:
      - "project"
      - "meeting"
      - "todo"
```

### Code Documentation

```yaml
tasks:
  - output_file: "api-docs.md"
    template: "templates/api-template.md"
    folders:
      - "docs/api"
    extensions:
      - ".md"
    tags:
      - "api"
      - "endpoint"

  - output_file: "tutorials.md"
    folders:
      - "docs/tutorials"
    extensions:
      - ".md"
    format: "📚 [%s](%s) - *Updated %s*\n"
```

### Multi-Extension Inventory

```yaml
tasks:
  - output_file: "all-code.md"
    folders:
      - "src"
      - "lib"
    extensions:
      - ".go"
      - ".js"
      - ".ts"
      - ".py"
    format: "- `%s` → [Source](%s) *(%s)*\n"
```

## Output Examples

### Default Format Output

```markdown
# AI Research Notes

Here are all my AI research documents.

## Research Papers

- [transformer-architecture](notes/transformer-architecture.md) 2024-12-15
- [attention-mechanisms](research/attention-mechanisms.md) 2024-12-10
- [gpt-analysis](papers/gpt-analysis.md) 2024-12-08
```

### Table Format Output

```markdown
# Documentation

| Document | Last Modified |
|----------|---------------|
| [api-reference](docs/api-reference.md) | 2024-12-15 |
| [user-guide](docs/user-guide.md) | 2024-12-10 |
| [installation](docs/installation.md) | 2024-12-08 |
```

## Use Cases

- **Documentation Management**: Keep track of all documentation files
- **Note Organization**: Organize notes by topic using tags
- **Diagram Inventories**: Catalog Excalidraw, Mermaid, or other diagram files
- **Code Documentation**: Generate indexes of code files, tests, or examples
- **Research Papers**: Organize academic papers and research notes
- **Project Tracking**: Create inventories of project-related files

## License

MIT License - see LICENSE file for details.

## Changelog

### v1.0.0
- Initial release
- Multi-task processing
- Tag filtering (YAML, TOML, inline)
- Custom formatting
- Template support
- Cross-platform compatibility
