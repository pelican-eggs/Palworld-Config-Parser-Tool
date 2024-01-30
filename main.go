package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"runtime"
)

func main() {
	// Read environment variables
	envVars := map[string]string{
		"ServerPlayerMaxNum": os.Getenv("MAX_PLAYERS"),
		"ServerName":         os.Getenv("SERVER_NAME"),
		"ServerPassword":     os.Getenv("SERVER_PASSWORD"),
		"AdminPassword":      os.Getenv("ADMIN_PASSWORD"),
		"PublicIP":           os.Getenv("PUBLIC_IP"),
		// Add other environment variables and corresponding INI keys here
	}

	// Specify keys for which quotes should be added
	envVarsQuotes := map[string]bool{
		"ServerName":     true,
		"ServerPassword": true,
		"AdminPassword":  true,
		// Add other keys as needed
	}

	// Determine the operating system
	var osFolder string
	switch runtime.GOOS {
	case "windows":
		osFolder = "WindowsServer"
	case "linux":
		osFolder = "LinuxServer"
	default:
		fmt.Println("Unsupported operating system")
		return
	}

	// Get the absolute path to the INI file
	iniFilePath, err := filepath.Abs(fmt.Sprintf("Pal/Saved/Config/%s/PalWorldSettings.ini", osFolder))
		if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}

	// Read the contents of the original INI file
	iniContent, err := ioutil.ReadFile(iniFilePath)
	if err != nil {
		fmt.Printf("Error reading INI file: %v\n", err)
		return
	}

	// Update values based on environment variables
	for key, value := range envVars {
		if value != "" {
			fmt.Printf("Updating key: %s with value: %s\n", key, value)
			setINIValue(&iniContent, key, value, envVarsQuotes[key])
		}
	}

	// Write the updated contents back to the INI file
	err = ioutil.WriteFile(iniFilePath, iniContent, 0644)
	if err != nil {
		fmt.Printf("Error writing updated INI file: %v\n", err)
		return
	}

	fmt.Println("INI file updated successfully.")
}

// setINIValue updates the value for the specified key in the INI content.
func setINIValue(content *[]byte, key, value string, addQuotes bool) {
	// Convert content to string for easy manipulation
	contentStr := string(*content)

	// Create the search string for the key
	searchStr := fmt.Sprintf("%s=", key)

	// Find the position of the key in the content
	pos := strings.Index(contentStr, searchStr)
	if pos == -1 {
		// Key not found
		fmt.Printf("Key not found: %s\n", key)
		return
	}

	// Find the end position of the value (comma or end of line)
	endPos := strings.Index(contentStr[pos:], ",")
	if endPos == -1 {
		// If there is no comma, check if it's at the end of the content
		endPos = len(contentStr) - pos
	}

	// If addQuotes is true and the key requires quotes, add quotes around the value
	if addQuotes {
		value = fmt.Sprintf(`"%s"`, value)
	}

	// Calculate the positions in the byte slice
	start := pos + len(searchStr)
	end := pos + endPos

	// Ensure the end position is within the slice bounds
	if end > len(*content) {
		end = len(*content)
	}

	// Update the content slice in place
	*content = append((*content)[:start], append([]byte(value), (*content)[end:]...)...)
}
