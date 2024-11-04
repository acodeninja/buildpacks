package fontconfig

import (
	"bufio"
	"fmt"
	"github.com/buildpacks/libcnb"
	"os"
	"path/filepath"
	"strings"
)

func ConfigPathRepoint(layer libcnb.Layer) error {
	// Open the input file for reading and writing
	file, err := os.OpenFile(fmt.Sprintf("%s/etc/fonts/fonts.conf", layer.Path), os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file content into memory
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		modifiedLine := line

		// Look for paths within XML-like tags (e.g., <dir>path</dir>)
		start := 0
		for {
			openTagPos := strings.Index(modifiedLine[start:], "<")
			closeTagPos := strings.Index(modifiedLine[start:], ">")
			endTagPos := strings.Index(modifiedLine[start:], "</")

			// Ensure tags are correctly found and positioned
			if openTagPos == -1 || closeTagPos == -1 || endTagPos == -1 || closeTagPos < openTagPos || endTagPos < closeTagPos {
				break
			}

			// Adjust positions relative to the full string
			openTagPos += start
			closeTagPos += start
			endTagPos += start

			// Extract path content between tags (e.g., the content of <dir>path</dir>)
			content := modifiedLine[closeTagPos+1 : endTagPos]

			// Check if the content is an absolute path
			if strings.HasPrefix(content, "/") && filepath.IsAbs(content) {
				// Add the prefix
				modifiedContent := layer.Path + content
				modifiedLine = modifiedLine[:closeTagPos+1] + modifiedContent + modifiedLine[endTagPos:]
			}

			// Move start to the next part of the line to avoid reprocessing
			start = endTagPos + 1
		}

		// Store the modified line
		lines = append(lines, modifiedLine)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Truncate the file to overwrite it with modified content
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to the beginning of the file: %w", err)
	}

	// Write modified content back to the file
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
