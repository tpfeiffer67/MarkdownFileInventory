package main

import (
	"os"
	"reflect"
	"testing"
)

func TestExtractTagsFromYAMLFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "No frontmatter",
			content:  "Just some regular content.",
			expected: nil,
		},
		{
			name: "Simple YAML frontmatter with one tag",
			content: `---
tags: tag1
---
Content here.`,
			expected: []string{"tag1"},
		},
		{
			name: "Simple YAML frontmatter with multiple tags",
			content: `---
tags: [tag1, tag2, tag3]
---
Content here.`,
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "YAML frontmatter with other keys",
			content: `---
title: My Doc
tags: [test]
author: me
---
Content here.`,
			expected: []string{"test"},
		},
		{
			name:     "Malformed YAML",
			content:  `--- tags: [test ---`,
			expected: nil,
		},
		{
			name: "YAML frontmatter with list format",
			content: `---
tags:
  - tagA
  - tagB
---
Content here.`,
			expected: []string{"tagA", "tagB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTagsFromYAMLFrontmatter(tt.content)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractTagsFromYAMLFrontmatter() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractTagsFromTOMLFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "No frontmatter",
			content:  "Just some regular content.",
			expected: nil,
		},
		{
			name: "Simple TOML frontmatter with one tag",
			content: `+++
tags = "tag1"
+++
Content here.`,
			expected: []string{"tag1"},
		},
		{
			name: "Simple TOML frontmatter with multiple tags",
			content: `+++
tags = ["tag1", "tag2", "tag3"]
+++
Content here.`,
			expected: []string{"tag1", "tag2", "tag3"},
		},
		{
			name: "TOML frontmatter with other keys",
			content: `+++
title = "My Doc"
tags = ["test"]
author = "me"
+++
Content here.`,
			expected: []string{"test"},
		},
		{
			name:     "Malformed TOML",
			content:  `+++ tags = ["test" +++`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTagsFromTOMLFrontmatter(tt.content)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractTagsFromTOMLFrontmatter() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFileContainsTags(t *testing.T) {
	tmpDir := t.TempDir()

	createTempFile := func(content string) string {
		file, err := os.CreateTemp(tmpDir, "testfile-*.md")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		_, err = file.WriteString(content)
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		file.Close()
		return file.Name()
	}

	yamlFile := createTempFile(`---
tags: [yaml_tag, shared]
---
Content with #inline_tag.`)

	tomlFile := createTempFile(`+++
tags = ["toml_tag", "shared"]
+++
Content.`)

	inlineFile := createTempFile(`Just content with #inline_tag and #another_tag.`)
	
	noTagFile := createTempFile(`No tags here.`)

	tests := []struct {
		name         string
		filePath     string
		requiredTags []string
		expected     bool
	}{
		{
			name:         "No required tags",
			filePath:     yamlFile,
			requiredTags: []string{},
			expected:     true,
		},
		{
			name:         "YAML tag found",
			filePath:     yamlFile,
			requiredTags: []string{"yaml_tag"},
			expected:     true,
		},
		{
			name:         "TOML tag found",
			filePath:     tomlFile,
			requiredTags: []string{"toml_tag"},
			expected:     true,
		},
		{
			name:         "Inline tag found",
			filePath:     inlineFile,
			requiredTags: []string{"inline_tag"},
			expected:     true,
		},
		{
			name:         "Inline tag found with hash",
			filePath:     inlineFile,
			requiredTags: []string{"#another_tag"},
			expected:     true,
		},
		{
			name:         "Shared tag in YAML",
			filePath:     yamlFile,
			requiredTags: []string{"shared"},
			expected:     true,
		},
		{
			name:         "Shared tag in TOML",
			filePath:     tomlFile,
			requiredTags: []string{"shared"},
			expected:     true,
		},
		{
			name:         "Tag not found",
			filePath:     noTagFile,
			requiredTags: []string{"nonexistent"},
			expected:     false,
		},
		{
			name:         "File with no tags, requires tags",
			filePath:     noTagFile,
			requiredTags: []string{"anytag"},
			expected:     false,
		},
		{
			name:         "One of many required tags found",
			filePath:     yamlFile,
			requiredTags: []string{"nonexistent", "yaml_tag"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileContainsTags(tt.filePath, tt.requiredTags)
			if got != tt.expected {
				t.Errorf("fileContainsTags() = %v, want %v", got, tt.expected)
			}
		})
	}
}
