package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// WeatherData - structure of the JSON response from OpenWeather API
type WeatherData struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temperature float64 `json:"temp"`
	} `json:"main"`
}

// getAPIKey - Fetch the OpenWeather API key
//
// ToDo: in a production environment we should be pulling this from a secret vault, not opsys env var.
// ToDo: validating the apiKey will have performance implications at scale, and pre-validating the source
//
//	may be the better solution.
func getAPIKey() (string, error) {
	const apiKeyRegex = "^[a-f0-9]{32}$"
	apiKey := strings.TrimSpace(os.Getenv("OPENWEATHER_API_KEY"))
	if apiKey == "" {
		return apiKey, fmt.Errorf("OPENWEATHER_API_KEY is not set")
	}
	pattern := regexp.MustCompile(apiKeyRegex)
	if !pattern.MatchString(apiKey) {
		return apiKey, fmt.Errorf("API key failed pattern check")
	} else {
		return apiKey, nil
	}
}

// validateLatitude - Verify that the given latitude is valid
// We don't want to pass unsanitized information to a vendor's API
func validateLatitude(raw string) (float64, error) {
	lat, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid latitude format: %s", raw)
	}
	if lat < -90 || lat > 90 {
		return 0, fmt.Errorf("latitude out of range (-90 to 90 degrees): %f", lat)
	}
	return lat, nil
}

// validateLongitude - Verify that the given longitude is valid
// We don't want to pass unsanitized information to a vendor's API
func validateLongitude(raw string) (float64, error) {
	lon, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid longitude format: %s", raw)
	}
	if lon < -180 || lon > 180 {
		return 0, fmt.Errorf("longitude out of range (-180 to 180 degrees): %f", lon)
	}
	return lon, nil
}

// healthCheck - provide a simple healthcheck response
func healthCheck(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("ok")); err != nil {
		log.Printf("healthcheck failed: %v", err)
	}
}

// weatherHandler - http handler
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := getAPIKey()
	if apiKey == "" {
		http.Error(w, "invalid API key", http.StatusInternalServerError)
		return
	}

	latitude, err := validateLatitude(r.URL.Query().Get("lat"))
	if err != nil {
		log.Printf("input error: %v", err)
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	longitude, err := validateLongitude(r.URL.Query().Get("lon"))
	if err != nil {
		log.Printf("input error: %v", err)
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	// Construct the API request URL
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&units=metric&appid=%s",
		latitude, longitude, apiKey)

	// Make the HTTP request to OpenWeather API
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("error closing body: %v", err)
		}
	}()

	var weatherData WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the weather condition & temperature information
	weatherCondition := weatherData.Weather[0].Description
	temperature := weatherData.Main.Temperature
	temperatureDesc := getTemperature(temperature)

	httpResponse := fmt.Sprintf("Current Temperature:\n"+
		"  Weather     : %s\n"+
		"  Temperature : %s", weatherCondition, temperatureDesc)

	// Send the response
	w.Header().Set("Content-Type", "text/plain")
	if _, err = fmt.Fprintf(w, httpResponse); err != nil {
		log.Printf("error writing the response: %v", err)
	}
}

// getTemperature - Given temperature (in Celsius), determine hot/cold
// I'm sure my European and Australian friends will appreciate this...
// But we'll convert it to Fahrenheit as well for grins.
func getTemperature(temp float64) string {
	if temp > 24 {
		return fmt.Sprintf("Hot (%.0fF / %.0fC)", celsiusToFahrenheit(temp), temp)
	} else if temp < 10 {
		return fmt.Sprintf("Cold (%.0fF / %.0fC)", celsiusToFahrenheit(temp), temp)
	} else {
		return fmt.Sprintf("Moderate (%.0fF / %.0fC)", celsiusToFahrenheit(temp), temp)
	}
}

// celsiusToFahrenheit - convert celsius to fahrenheit
func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 9.0 / 5.0) + 32.0
}

// GetHttpListenAddressAndPort - Get the IP addr and port we will listen on
// Verify that the address and port are valid.
func GetHttpListenAddressAndPort() (string, error) {
	rawAddr := os.Getenv("HTTP_LISTEN_ADDR")
	rawPort := os.Getenv("HTTP_LISTEN_PORT")

	if strings.TrimSpace(rawAddr) == "" {
		return "", fmt.Errorf("missing IP address (HTTP_LISTEN_ADDR not set)")
	}

	if strings.TrimSpace(rawPort) == "" {
		return "", fmt.Errorf("missing port (HTTP_LISTEN_PORT not set)")
	}

	// Verify rawAddr is a valid IP address
	ip := net.ParseIP(rawAddr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", rawAddr)
	}

	// Verify rawPort is a valid port number
	port, err := strconv.Atoi(rawPort)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid port number: %s", rawPort)
	}

	return fmt.Sprintf("%s:%d", rawAddr, port), nil
}

func main() {

	listenAddress, err := GetHttpListenAddressAndPort()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/weather", weatherHandler)
	fmt.Printf("Server listening on port %s...\n", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
