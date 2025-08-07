# Texas License Search

A simple Go command-line tool to search and retrieve professional license data from the Texas Open Data Portal. All search parameters support **partial matching** and are **case-insensitive** for maximum flexibility.

## Overview

This tool queries the Texas government's public API to fetch information about various professional licenses issued in Texas. It supports flexible filtering with partial, case-insensitive matching across multiple fields including license type, business county, license subtype, business name, owner name, license number, and expiration date.

## Prerequisites

- Go 1.16 or higher
- A Texas Open Data Portal App Token (free registration required)

## Installation

1. Clone or download this repository
2. Install dependencies:
   ```bash
   go mod init texas-license-fetcher
   go mod tidy
   ```

## Setup

### Get an App Token

1. Visit [data.texas.gov](https://data.texas.gov)
2. Create a free account
3. Generate an App Token from your account settings
4. Set the token as an environment variable:
   ```bash
   export APP_TOKEN=your_app_token_here
   ```

## Usage

### Basic Usage

```bash
go run main.go [flags]
```

### Available Flags

| Flag | Description | Example |
|------|-------------|---------|
| `-e` | License expiration date (partial match) | `-e "2025"` or `-e "12/16"` |
| `-n` | License number (partial match) | `-n "90210"` or `-n "TACLA"` |
| `-t` | License type (partial, case-insensitive) | `-t "plumb"` finds "Plumber" |
| `-c` | Business county (partial, case-insensitive) | `-c "harris"` finds "HARRIS" |
| `-st` | License subtype (partial, case-insensitive) | `-st "reg"` finds "REG" |
| `-bn` | Business name (partial, case-insensitive) | `-bn "bob"` finds "BOB'S PLUMBING" |
| `-on` | Owner name (partial, case-insensitive) | `-on "smith"` finds "SMITH, JOHN" |
| `-timeout` | HTTP request timeout in seconds | `-timeout 60` |
| `-limit` | Maximum records to retrieve | `-limit 100` |

### Examples

**Search for all A/C Technicians in Harris County:**
```bash
go run main.go -t "a/c" -c "harris"
```

**Find licenses for businesses with "plumbing" in the name:**
```bash
go run main.go -bn "plumbing" -limit 10
```

**Search by partial owner name:**
```bash
go run main.go -on "smith" -limit 50
```

**Find licenses expiring in 2025:**
```bash
go run main.go -e "2025" -limit 20
```

**Search for specific license numbers:**
```bash
go run main.go -n "TACLA" -limit 10
```

**Combine multiple search criteria:**
```bash
go run main.go -t "plumb" -c "harris" -on "bob" -limit 25
```

## Search Features

- **Partial Matching**: All text searches support partial matches - searching for "plumb" will find "Plumber", "Plumbing", etc.
- **Case-Insensitive**: Search terms are automatically converted to match the database format - "harris" finds "HARRIS"
- **Multiple Criteria**: Combine any search parameters to narrow results
- **Flexible Date Search**: Search expiration dates by year, month/year, or full date
- **License Number Search**: Find licenses by partial license numbers or prefixes

## Output

The tool outputs each license record as formatted JSON, followed by a total count. Example output:

```json
{
  "license_type": "A/C Technician",
  "license_number": "TACLA12345",
  "business_county": "HARRIS",
  "business_name": "Cool Air Solutions",
  "owner_name": "John Doe",
  "mailing_address_county": "HARRIS",
  "license_expiration_date_mmddccyy": "12/31/2024",
  "license_subtype": "REG",
  "continuing_education_flag": "Y"
}
```

## Features

- **Pagination Support**: Automatically handles large result sets by fetching data in chunks
- **Graceful Shutdown**: Responds to Ctrl+C (SIGINT) and SIGTERM signals
- **Formatted Output**: Pretty-printed, colorized JSON output for easy reading
- **Flexible Search**: Partial, case-insensitive matching across all text fields
- **Result Limiting**: Control the maximum number of records retrieved
- **Configurable Timeout**: Set custom HTTP request timeouts

## Error Handling

The tool will exit with an error message if:
- The `APP_TOKEN` environment variable is not set
- The API returns a non-200 status code
- Network connectivity issues occur
- Invalid JSON is returned from the API

## Data Source

This tool queries the Texas Professional Licensing dataset available at:
- **API Endpoint**: `https://data.texas.gov/resource/7358-krk7.json`
- **Dataset**: Professional Licensing Data
- **Update Frequency**: Regular updates from Texas state agencies

## Dependencies

- `github.com/tidwall/pretty` - For JSON formatting and colorization

## License

This project is provided as-is for educational and informational purposes. Please respect the Texas Open Data Portal's terms of service and rate limits.

## Troubleshooting

**"Didn't find required APP_TOKEN in env"**
- Make sure you've set the `APP_TOKEN` environment variable with your Texas Open Data Portal token

**"API returned a non-200 status code"**
- Check your internet connection
- Verify your App Token is valid
- Ensure your search parameters are correctly formatted

**No results returned**
- Try broadening your search criteria - remember all searches are partial matches
- Check for typos in your search terms
- Try searching with fewer criteria to see if records exist
- For license numbers, try searching with just the prefix (e.g., "TACLA" instead of full number)
