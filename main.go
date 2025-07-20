package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type FileInfo struct {
	Path    string
	ModTime time.Time
}

type Task struct {
	OutputFile string   `yaml:"output_file"`
	Template   string   `yaml:"template"`
	Folders    []string `yaml:"folders"`
	Extensions []string `yaml:"extensions"`
	Tags       []string `yaml:"tags"`
	Format     string   `yaml:"format"`
}

type Config struct {
	Tasks []Task `yaml:"tasks"`
}

type FrontMatterTOML struct {
	Tags interface{} `toml:"tags"`
}

type FrontMatterYAML struct {
	Tags interface{} `yaml:"tags"`
}

// Extract tags from YAML frontmatter
func extractTagsFromYAMLFrontmatter(content string) []string {
	var tags []string

	// Look for YAML frontmatter delimited by ---
	if strings.HasPrefix(content, "---") {
		endIndex := strings.Index(content[3:], "---")
		if endIndex != -1 {
			frontmatterContent := content[3 : endIndex+3]

			var fm FrontMatterYAML
			if err := yaml.Unmarshal([]byte(frontmatterContent), &fm); err == nil {
				switch v := fm.Tags.(type) {
				case string:
					// Single tag: tags: ia
					tags = append(tags, v)
				case []interface{}:
					// Multiple tags: tags: [ia, prompt]
					for _, tag := range v {
						if str, ok := tag.(string); ok {
							tags = append(tags, str)
						}
					}
				}
			}
		}
	}

	return tags
}

// Extract tags from TOML frontmatter
func extractTagsFromTOMLFrontmatter(content string) []string {
	var tags []string

	// Look for TOML frontmatter delimited by +++
	if strings.HasPrefix(content, "+++") {
		endIndex := strings.Index(content[3:], "+++")
		if endIndex != -1 {
			frontmatterContent := content[3 : endIndex+3]

			var fm FrontMatterTOML
			if err := toml.Unmarshal([]byte(frontmatterContent), &fm); err == nil {
				switch v := fm.Tags.(type) {
				case string:
					// Single tag: tags = "ia"
					tags = append(tags, v)
				case []interface{}:
					// Multiple tags: tags = ["ia", "prompt"]
					for _, tag := range v {
						if str, ok := tag.(string); ok {
							tags = append(tags, str)
						}
					}
				}
			}
		}
	}

	return tags
}

// Check if file contains any of the specified tags
func fileContainsTags(filePath string, requiredTags []string) bool {
	// If no tags specified, include all files
	if len(requiredTags) == 0 {
		return true
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		// If we can't read the file, exclude it
		return false
	}

	contentStr := string(content)

	// Check tags in YAML frontmatter (---)
	yamlTags := extractTagsFromYAMLFrontmatter(contentStr)
	for _, frontTag := range yamlTags {
		for _, reqTag := range requiredTags {
			// Remove # from required tag if present for frontmatter comparison
			cleanReqTag := strings.TrimPrefix(reqTag, "#")
			if frontTag == cleanReqTag {
				return true
			}
		}
	}

	// Check tags in TOML frontmatter (+++)
	tomlTags := extractTagsFromTOMLFrontmatter(contentStr)
	for _, frontTag := range tomlTags {
		for _, reqTag := range requiredTags {
			// Remove # from required tag if present for frontmatter comparison
			cleanReqTag := strings.TrimPrefix(reqTag, "#")
			if frontTag == cleanReqTag {
				return true
			}
		}
	}

	// Check if any tag is present in the file content (with #)
	for _, tag := range requiredTags {
		// Ensure tag starts with # for content search
		searchTag := tag
		if !strings.HasPrefix(searchTag, "#") {
			searchTag = "#" + tag
		}
		if strings.Contains(contentStr, searchTag) {
			return true
		}
	}

	return false
}

func processAllTasks(rootFolder string, config Config) {
	for i, task := range config.Tasks {
		if err := processTask(task, rootFolder, i+1); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing task %d (%s): %v\n", i+1, task.OutputFile, err)
		}
	}
	fmt.Printf("All tasks completed successfully!\n")
}

func processTask(task Task, rootFolder string, taskNum int) error {
	tagsInfo := ""
	if len(task.Tags) > 0 {
		tagsInfo = fmt.Sprintf(" (filtering by tags: %s)", strings.Join(task.Tags, ", "))
	}
	fmt.Printf("Processing task %d: %s%s\n", taskNum, task.OutputFile, tagsInfo)

	var allFiles []FileInfo
	for _, folder := range task.Folders {
		folderPath := filepath.Join(rootFolder, folder)
		for _, extension := range task.Extensions {
			if !strings.HasPrefix(extension, ".") {
				extension = "." + extension
			}

			walkErr := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Cannot access path %s: %v. Skipping.\n", path, err)
					return nil
				}
				if info.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, extension) {
					if fileContainsTags(path, task.Tags) {
						relPath, err := filepath.Rel(rootFolder, path)
						if err != nil {
							relPath = path
						}
						allFiles = append(allFiles, FileInfo{
							Path:    relPath,
							ModTime: info.ModTime(),
						})
					}
				}
				return nil
			})

			if walkErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Error walking folder %s: %v\n", folderPath, walkErr)
			}
		}
	}

	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Path < allFiles[j].Path
	})

	outputPath := filepath.Join(rootFolder, task.OutputFile)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", outputPath, err)
	}
	defer file.Close()

	if task.Template != "" {
		templatePath := filepath.Join(rootFolder, task.Template)
		templateContent, err := os.ReadFile(templatePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error reading template %s: %v\n", templatePath, err)
		} else {
			_, err = file.Write(templateContent)
			if err != nil {
				return fmt.Errorf("error writing template content to %s: %w", outputPath, err)
			}
			if !strings.HasSuffix(string(templateContent), "\n") {
				fmt.Fprint(file, "\n")
			}
		}
	}

	for _, fileInfo := range allFiles {
		fileName := filepath.Base(fileInfo.Path)
		displayName := fileName

		for _, ext := range task.Extensions {
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			if strings.HasSuffix(fileName, ext) {
				displayName = strings.TrimSuffix(fileName, ext)
				break
			}
		}

		markdownPath := strings.ReplaceAll(fileInfo.Path, "\\", "/")
		encodedPath := strings.ReplaceAll(markdownPath, " ", "%20")
		modDate := fileInfo.ModTime.Format("2006-01-02")

		format := task.Format
		if format == "" {
			format = "- [%s](%s) %s\n"
		}

		_, err := fmt.Fprintf(file, format, displayName, encodedPath, modDate)
		if err != nil {
			return fmt.Errorf("error writing to output file %s: %w", outputPath, err)
		}
	}

	templateMsg := ""
	if task.Template != "" {
		templateMsg = fmt.Sprintf(" (using template: %s)", task.Template)
	}
	fmt.Printf("  Created %s with %d files%s\n", task.OutputFile, len(allFiles), templateMsg)
	return nil
}

func watchMode(rootFolder string, config Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("Modified file:", event.Name, "- Re-running tasks.")
					processAllTasks(rootFolder, config)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add all folders from all tasks to the watcher
	for _, task := range config.Tasks {
		for _, folder := range task.Folders {
			folderPath := filepath.Join(rootFolder, folder)
			err = watcher.Add(folderPath)
			if err != nil {
				log.Printf("Error watching folder %s: %v", folderPath, err)
			}
		}
	}

	log.Println("Watching for changes. Press Ctrl+C to exit.")
	<-done
}

func main() {
	watch := flag.Bool("watch", false, "Enable watch mode to automatically re-run tasks on file changes.")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--watch] <root_folder>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s ./project\n", os.Args[0])
		os.Exit(1)
	}

	rootFolder := flag.Arg(0)
	configPath := filepath.Join(rootFolder, ".markdown-file-inventory.yaml")

	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file %s: %v\n", configPath, err)
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML config: %v\n", err)
		os.Exit(1)
	}

	processAllTasks(rootFolder, config)

	if *watch {
		watchMode(rootFolder, config)
	}
}
