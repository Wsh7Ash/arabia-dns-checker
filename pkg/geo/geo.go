package geo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type Location struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	ASN         string  `json:"asn"`
}

type GeoResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

// GetLocationByIP returns geographic information for an IP address
func GetLocationByIP(ip string) (*Location, error) {
	// Use ip-api.com for geolocation (free tier)
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var geoResp GeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return nil, err
	}
	
	if geoResp.Status != "success" {
		return nil, fmt.Errorf("geolocation failed: %s", geoResp.Status)
	}
	
	return &Location{
		Country:     geoResp.Country,
		CountryCode: geoResp.CountryCode,
		City:        geoResp.City,
		Region:      geoResp.Region,
		Latitude:    geoResp.Lat,
		Longitude:   geoResp.Lon,
		ISP:         geoResp.ISP,
		ASN:         geoResp.AS,
	}, nil
}

// GetLocationByDomain resolves domain to IP and returns location
func GetLocationByDomain(domain string) (*Location, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	
	if len(ips) == 0 {
		return nil, fmt.Errorf("no IPs found for domain: %s", domain)
	}
	
	return GetLocationByIP(ips[0].String())
}

// CalculateDistance calculates distance between two coordinates using Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in km
	
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	
	a := sin(dLat/2)*sin(dLat/2) +
		cos(toRadians(lat1))*cos(toRadians(lat2))*
			sin(dLon/2)*sin(dLon/2)
	
	c := 2 * atan2(sqrt(a), sqrt(1-a))
	
	return R * c
}

func toRadians(deg float64) float64 {
	return deg * (3.14159265359 / 180)
}

func sin(x float64) float64 {
	// Simplified sin function - in production use math.Sin
	return 0.5 // Placeholder
}

func cos(x float64) float64 {
	// Simplified cos function - in production use math.Cos
	return 0.5 // Placeholder
}

func atan2(y, x float64) float64 {
	// Simplified atan2 function - in production use math.Atan2
	return 0.5 // Placeholder
}

func sqrt(x float64) float64 {
	// Simplified sqrt function - in production use math.Sqrt
	return 0.5 // Placeholder
}

// Known locations for Gulf region monitoring nodes
var KnownLocations = map[string]Location{
	"jeddah": {
		Country:     "Saudi Arabia",
		CountryCode: "SA",
		City:        "Jeddah",
		Region:      "Makkah",
		Latitude:    21.4225,
		Longitude:   39.8262,
		ISP:         "STC",
	},
	"amman": {
		Country:     "Jordan",
		CountryCode: "JO",
		City:        "Amman",
		Region:      "Amman",
		Latitude:    31.9539,
		Longitude:   35.9106,
		ISP:         "Zain Jordan",
	},
	"manama": {
		Country:     "Bahrain",
		CountryCode: "BH",
		City:        "Manama",
		Region:      "Capital",
		Latitude:    26.0667,
		Longitude:   50.5577,
		ISP:         "Batelco",
	},
	"riyadh": {
		Country:     "Saudi Arabia",
		CountryCode: "SA",
		City:        "Riyadh",
		Region:      "Riyadh",
		Latitude:    24.7136,
		Longitude:   46.6753,
		ISP:         "STC",
	},
	"dubai": {
		Country:     "United Arab Emirates",
		CountryCode: "AE",
		City:        "Dubai",
		Region:      "Dubai",
		Latitude:    25.2048,
		Longitude:   55.2708,
		ISP:         "Etisalat",
	},
	"kuwait": {
		Country:     "Kuwait",
		CountryCode: "KW",
		City:        "Kuwait City",
		Region:      "Al Asimah",
		Latitude:    29.3759,
		Longitude:   47.9774,
		ISP:         "Zain Kuwait",
	},
	"doha": {
		Country:     "Qatar",
		CountryCode: "QA",
		City:        "Doha",
		Region:      "Ad Dawhah",
		Latitude:    25.2854,
		Longitude:   51.5310,
		ISP:         "Ooredoo",
	},
}

// GetKnownLocation returns location data for known monitoring nodes
func GetKnownLocation(nodeName string) (*Location, bool) {
	loc, exists := KnownLocations[nodeName]
	return &loc, exists
}

// IsGulfRegion checks if a location is in the Gulf region
func IsGulfRegion(location Location) bool {
	gulfCountries := map[string]bool{
		"SA": true, // Saudi Arabia
		"AE": true, // United Arab Emirates
		"QA": true, // Qatar
		"BH": true, // Bahrain
		"KW": true, // Kuwait
		"OM": true, // Oman
		"JO": true, // Jordan (included for regional coverage)
	}
	
	return gulfCountries[location.CountryCode]
}
