package main

import (
	"fmt"
	"math"
	"os"
	"testing"
)

func TestGetHttpListenAddressAndPort(t *testing.T) {

	t.Run("Missing environment variables (all)", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
		})
		// make sure we don't accidentally have something set
		_ = os.Unsetenv("HTTP_LISTEN_ADDR")
		_ = os.Unsetenv("HTTP_LISTEN_PORT")
		_, err := GetHttpListenAddressAndPort()
		if err == nil {
			t.Fatalf("expected missing ip error. got none")
		}
		if err.Error() != "missing IP address (HTTP_LISTEN_ADDR not set)" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Missing environment variables (only port)", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
		})
		_ = os.Setenv("HTTP_LISTEN_ADDR", "127.0.0.1")
		_ = os.Unsetenv("HTTP_LISTEN_PORT")
		_, err := GetHttpListenAddressAndPort()
		if err == nil {
			t.Fatalf("expected missing port error. got none")
		}
		if err.Error() != "missing port (HTTP_LISTEN_PORT not set)" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Valid environment variables", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
		})
		_ = os.Setenv("HTTP_LISTEN_ADDR", "127.0.0.1")
		_ = os.Setenv("HTTP_LISTEN_PORT", "8080")
		addr, err := GetHttpListenAddressAndPort()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if addr != "127.0.0.1:8080" {
			t.Errorf("Expected address '127.0.0.1:8080', got '%s'", addr)
		}
	})

	t.Run("Invalid IP address", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
		})
		_ = os.Setenv("HTTP_LISTEN_ADDR", "invalid_address")
		_ = os.Setenv("HTTP_LISTEN_PORT", "8080")
		_, err := GetHttpListenAddressAndPort()
		if err == nil {
			t.Error("Expected error for invalid IP address")
		}
	})

	t.Run("Invalid port number", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
		})
		_ = os.Setenv("HTTP_LISTEN_ADDR", "127.0.0.1")
		_ = os.Setenv("HTTP_LISTEN_PORT", "invalid_port")
		_, err := GetHttpListenAddressAndPort()
		if err == nil {
			t.Error("Expected error for invalid port number")
		}
	})

}

func TestGetApiKey(t *testing.T) {

	t.Run("unset ApiKey.  Expect error", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
		})
		_ = os.Setenv("HTTP_LISTEN_ADDR", "127.0.0.1")
		_ = os.Setenv("HTTP_LISTEN_PORT", "8080")
		_ = os.Unsetenv("OPENWEATHER_API_KEY")
		_, err := getAPIKey()
		if err == nil {
			t.Fatalf("Expected error.  got none.")
		}
		if err.Error() != "OPENWEATHER_API_KEY is not set" {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("Valid API key", func(t *testing.T) {
		const fakeApiKey = "abcdef0123456789abcdef0123456789"
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
			_ = os.Unsetenv("OPENWEATHER_API_KEY")
		})
		_ = os.Setenv("HTTP_LISTEN_ADDR", "127.0.0.1")
		_ = os.Setenv("HTTP_LISTEN_PORT", "8080")
		_ = os.Setenv("OPENWEATHER_API_KEY", fakeApiKey)

		_ = os.Setenv("OPENWEATHER_API_KEY", fakeApiKey)
		apiKey, err := getAPIKey()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if apiKey != fakeApiKey {
			t.Fatalf("Expected API key '%s', got '%s'", fakeApiKey, apiKey)
		}
	})

	t.Run("Invalid API key (not matching regex)", func(t *testing.T) {
		t.Cleanup(func() {
			// Clean up environment variables
			_ = os.Unsetenv("HTTP_LISTEN_ADDR")
			_ = os.Unsetenv("HTTP_LISTEN_PORT")
			_ = os.Unsetenv("OPENWEATHER_API_KEY")
		})
		_ = os.Setenv("OPENWEATHER_API_KEY", "def0123456789abcdef0123456789xyz")
		_, err := getAPIKey()
		if err == nil {
			t.Fatalf("Expected error for invalid API key")
		}
	})
}

func TestValidateLatitude(t *testing.T) {
	t.Run("Valid latitude within range", func(t *testing.T) {
		latStr := "37.7749"
		expectedLat := 37.7749
		lat, err := validateLatitude(latStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if lat != expectedLat {
			t.Errorf("Expected latitude %f, got %f", expectedLat, lat)
		}
	})

	t.Run("Latitude exactly at lower boundary", func(t *testing.T) {
		latStr := "-90.0"
		expectedLat := -90.0
		lat, err := validateLatitude(latStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if lat != expectedLat {
			t.Errorf("Expected latitude %f, got %f", expectedLat, lat)
		}
	})

	t.Run("Latitude exactly at upper boundary", func(t *testing.T) {
		latStr := "90.0"
		expectedLat := 90.0
		lat, err := validateLatitude(latStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if lat != expectedLat {
			t.Errorf("Expected latitude %f, got %f", expectedLat, lat)
		}
	})

	t.Run("Invalid latitude (out of bounds)", func(t *testing.T) {
		latStr := "100.0"
		_, err := validateLatitude(latStr)
		if err == nil {
			t.Error("Expected error for latitude out of range")
		}
	})

	t.Run("Invalid (non-numeric) latitude", func(t *testing.T) {
		latStr := "invalid_latitude"
		_, err := validateLatitude(latStr)
		if err == nil {
			t.Error("Expected error for invalid latitude format")
		}
	})
}

func TestValidateLongitude(t *testing.T) {
	t.Run("Valid Longitude within range", func(t *testing.T) {
		longStr := "37.7749"
		expectedLong := 37.7749
		longitude, err := validateLongitude(longStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if longitude != expectedLong {
			t.Errorf("Expected Longitude %f, got %f", expectedLong, longitude)
		}
	})

	t.Run("Longitude exactly at lower boundary", func(t *testing.T) {
		longStr := "-90.0"
		expectedLong := -90.0
		longitude, err := validateLongitude(longStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if longitude != expectedLong {
			t.Errorf("Expected Longitude %f, got %f", expectedLong, longitude)
		}
	})

	t.Run("Longitude exactly at upper boundary", func(t *testing.T) {
		longStr := "90.0"
		expectedLong := 90.0
		longitude, err := validateLongitude(longStr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if longitude != expectedLong {
			t.Errorf("Expected Longitude %f, got %f", expectedLong, longitude)
		}
	})

	t.Run("Invalid Longitude (out of bounds)", func(t *testing.T) {
		for _, n := range []float64{-200, -181, +200, +180.1} {
			invalidLongStr := fmt.Sprintf("%f", n)
			_, err := validateLongitude(invalidLongStr)
			if err == nil {
				t.Error("Expected error for Longitude out of range")
			}
		}
	})

	t.Run("Invalid (non-numeric) Longitude", func(t *testing.T) {
		longStr := "non-numeric-longitude"
		_, err := validateLongitude(longStr)
		if err == nil {
			t.Error("Expected error for invalid Longitude format")
		}
	})
}

func TestGetTemperature(t *testing.T) {
	testCases := []struct {
		temp     float64
		expected string
	}{
		{25.0, "Hot (77F / 25C)"},
		{15.0, "Moderate (59F / 15C)"},
		{5.0, "Cold (41F / 5C)"},
		{-5.0, "Cold (23F / -5C)"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Temperature %f", tc.temp), func(t *testing.T) {
			result := getTemperature(tc.temp)
			if result != tc.expected {
				t.Errorf("value mismatch\n"+
					"    Temp:  %f\n"+
					"Expected: '%s'\n"+
					"  Actual: '%s'", tc.temp, tc.expected, result)
			}
		})
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	testCases := []struct {
		celsius  float64
		expected float64
	}{
		{0.0, 32.0},        // Freezing point of water
		{100.0, 212.0},     // Boiling point of water
		{-40.0, -40.0},     // -40 degrees Celsius is -40 degrees Fahrenheit
		{37.0, 98.6},       // Normal body temperature in Fahrenheit
		{-273.15, -459.67}, // Absolute zero in Celsius to Fahrenheit
	}

	// Define a small tolerance for floating-point comparisons
	tolerance := 0.001 // Adjust as needed based on precision requirements

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Celsius %.2f", tc.celsius), func(t *testing.T) {
			result := celsiusToFahrenheit(tc.celsius)
			if math.Abs(result-tc.expected) > tolerance {
				t.Errorf("Expected %.2f°F, but got %.2f°F", tc.expected, result)
			}
		})
	}
}
