package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// assignmentData represents the structure of assignment files
type assignmentData struct {
	Order            uint32 `yaml:"order" json:"order"`
	Name             string `yaml:"name" json:"name,omitempty"`
	Title            string `yaml:"title" json:"title,omitempty"`
	Deadline         string `yaml:"deadline" json:"deadline"`
	IsGroupLab       bool   `yaml:"isgrouplab" json:"isgrouplab"`
	AutoApprove      bool   `yaml:"autoapprove" json:"autoapprove"`
	ScoreLimit       uint32 `yaml:"scorelimit" json:"scorelimit"`
	Reviewers        uint32 `yaml:"reviewers" json:"reviewers"`
	ContainerTimeout uint32 `yaml:"containertimeout" json:"containertimeout"`
	HoursMin         int    `yaml:"hoursmin" json:"hoursmin,omitempty"`
	HoursMax         int    `yaml:"hoursmax" json:"hoursmax,omitempty"`
}

// convertAssignments converts YAML assignment files to JSON format
func convertAssignments(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: cm convert-assignments <directory>")
		fmt.Println("Converts assignment.yml files to assignment.json files in the specified directory and subdirectories")
		return
	}

	dir := args[0]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Error: directory %s does not exist\n", dir)
		return
	}

	converted := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		if filename == "assignment.yml" || filename == "assignment.yaml" {
			if err := convertAssignmentFile(path); err != nil {
				fmt.Printf("Error converting %s: %v\n", path, err)
				return err
			}
			converted++
			fmt.Printf("Converted: %s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	fmt.Printf("Successfully converted %d assignment files\n", converted)
}

// convertAssignmentFile converts a single YAML assignment file to JSON
func convertAssignmentFile(yamlPath string) error {
	// Read YAML file
	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse YAML
	var assignment assignmentData
	if err := yaml.Unmarshal(yamlContent, &assignment); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to JSON
	jsonContent, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create JSON file path
	dir := filepath.Dir(yamlPath)
	jsonPath := filepath.Join(dir, "assignment.json")

	// Write JSON file
	if err := os.WriteFile(jsonPath, jsonContent, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}