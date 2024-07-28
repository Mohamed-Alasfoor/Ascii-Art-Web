package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Main function - entry point of the application
func main() {
	// Set up URL routes to their corresponding handlers.
	http.HandleFunc("/", Serverouter)

	// Start an HTTP server listening on port 8080.
	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

// Serverouter handles routing for different URL paths
func Serverouter(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		serveHome(w, r)
	case "/ascii-art":
		asciiArtHandler(w, r)
	case "/style.css":
		serveCSS(w, r)
	default:
		// Redirect to home page for any undefined routes
		renderError(w, "Method not allowed", http.StatusMethodNotAllowed) // 405 status code
	}
}

// serveHome handles requests for the home page
func serveHome(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is GET
	if r.Method != "GET" {
		renderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse and execute the home template
	tmpl, err := template.ParseFiles("HTML/home.html")
	if err != nil {
		renderError(w, "Internal Server Error: Failed to load template", http.StatusInternalServerError)
		return
	}
	// Handle any errors that occur during template parsing or execution
	err = tmpl.Execute(w, map[string]string{"Result": ""})
	if err != nil {
		renderError(w, "Internal Server Error: Failed to render template", http.StatusInternalServerError)
		return
	}
}

// serveCSS handles requests for the CSS file
func serveCSS(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is GET
	if r.Method != "GET" {
		renderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Serve the CSS file
	path := filepath.Join(".", "style.css")
	http.ServeFile(w, r, path)
}

// asciiArtHandler processes requests for ASCII art generation
func asciiArtHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != "POST" {
		renderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse form data and validate input
	if err := r.ParseForm(); err != nil {
		renderError(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")
	banner := r.FormValue("banner")
	if text == "" {
		renderError(w, "Missing text: please provide the text for ASCII art generation.", http.StatusBadRequest)
		return
	}
	if banner == "" {
		renderError(w, "Missing banner: please select a banner for ASCII art generation.", http.StatusBadRequest)
		return
	}

	// Open the banner file and generate ASCII art
	filePath := fmt.Sprintf("ART/%s.txt", banner)
	content, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			renderError(w, "Banner file not found", http.StatusNotFound)
		} else {
			renderError(w, "Internal Server Error: Failed to open banner file", http.StatusInternalServerError)
		}
		return
	}
	defer content.Close()

	result := generateASCIIArt(content, strings.Split(text, "\n"))
	// Render the result using the home template
	tmpl, err := template.ParseFiles("HTML/home.html")
	if err != nil {
		renderError(w, "Internal Server Error: Failed to load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, map[string]string{"Result": result})
	if err != nil {
		renderError(w, "Internal Server Error: Failed to render template", http.StatusInternalServerError)
		return
	}
}

// generateASCIIArt creates ASCII art from user input and a banner file
func generateASCIIArt(content *os.File, userInput []string) string {
	// Create a map to hold the ASCII representation of each character.
	asciiArtMap := make(map[rune][]string)

	// Assume each character's art is 8 lines high.
	const height = 8

	// Read the banner font characters into the map.
	scanner := bufio.NewScanner(content)
	for i := 32; i <= 126; i++ { // For all printable ASCII characters
		asciiArt := make([]string, height)
		for j := range asciiArt {
			if !scanner.Scan() {
				log.Fatal("Error reading banner font file")
			}
			asciiArt[j] = scanner.Text()
		}
		asciiArtMap[rune(i)] = asciiArt
		// Skip the blank line after each character's art
		if !scanner.Scan() {
			log.Fatal("Error reading banner font file")
		}
	}

	// Build the ASCII art for the user's input
	var result strings.Builder
	for _, line := range userInput {
		for i := 0; i < height; i++ {
			for _, char := range line {
				if art, ok := asciiArtMap[char]; ok {
					result.WriteString(art[i])
				} else {
					result.WriteString(" ") // Handle unknown characters
				}
			}
			result.WriteString("\n")
		}
		result.WriteString("\n") // Add an additional newline to separate the lines
	}

	return result.String()
}

// renderError displays an error message to the user
func renderError(w http.ResponseWriter, errMsg string, statusCode int) {
	// Set the HTTP status code
	w.WriteHeader(statusCode)
	// Parse and execute the error template
	tmpl, err := template.ParseFiles("HTML/error.html")
	if err != nil {
		// If the error template fails to load, log the error and send a basic error message
		log.Printf("Error loading error template: %v", err)
		http.Error(w, "500 Internal Server Error: Failed to load error template", http.StatusInternalServerError)
		return
	}
	// Handle any errors that occur during template parsing or execution
	tmpl.Execute(w, map[string]string{"ErrorMessage": errMsg})
}
