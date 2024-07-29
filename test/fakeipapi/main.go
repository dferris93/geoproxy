package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "strings"
)

// Response struct to map the JSON payload
type Response struct {
    Status      string `json:"status"`
    CountryCode string `json:"countryCode"`
    Region      string `json:"region"`
}

// Global variable to hold the country code and region
var response Response

func handleJSON(w http.ResponseWriter, r *http.Request) {
    // Split the URL path
    parts := strings.Split(r.URL.Path, "/")

    // Expecting path to be /json/<ip address>, so we need at least 3 parts
    if len(parts) < 3 {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Extract the IP address part
    ip := parts[2]

    // Log the IP address for demonstration purposes (not used in the response)
    log.Printf("IP Address: %s\n", ip)

    // Set the content type to application/json
    w.Header().Set("Content-Type", "application/json")

    // Encode the Response struct to JSON and send it as the response
    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
    }
}

func main() {
    // Define command line flags
    countryCode := flag.String("countryCode", "US", "a string")
    region := flag.String("region", "WA", "a string")
    flag.Parse()

    // Initialize the response struct with command line arguments
    response = Response{
        Status:      "success",
        CountryCode: *countryCode,
        Region:      *region,
    }

    // Define a handler for the /json/ route
    http.HandleFunc("/json/", handleJSON)

    // Start the web server on port 8080
    fmt.Println("Server is running on http://localhost:8181")
    log.Fatal(http.ListenAndServe("127.0.0.1:8181", nil))
}
