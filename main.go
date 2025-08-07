package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tidwall/pretty"
)

const (
	apiBaseURL = "https://data.texas.gov/resource/7358-krk7.json"
)

func main() {
	appToken := os.Getenv("APP_TOKEN")
	if appToken == "" {
		log.Fatal("Didn't find required APP_TOKEN in env")
	}

	expDate := flag.String("e", "", "The expiration date (eg. 12/16/2025)")
	licenseNumber := flag.String("n", "", "The license number (eg. 90210)")
	licenseType := flag.String("t", "", "The license type to search for (eg. A/C Technician)")
	businessCounty := flag.String("c", "", "The business county (eg. HARRIS)")
	licenseSubType := flag.String("st", "", "The license sub-type (eg. REG)")
	businessName := flag.String("bn", "", "The business name (eg. BOB'S PLUMBING)")
	ownerName := flag.String("on", "", "The owner name (eg. BOBS, BOBBY)")
	timeOutSecs := flag.Int("timeout", 30, "The timeout in seconds")
	var limit int
	flag.IntVar(&limit, "limit", 0, "The max records to retrieve")
	flag.Parse()

	recordsPerRequest := 5000
	if limit > 0 && limit < recordsPerRequest {
		recordsPerRequest = limit
	}

	records := make(chan json.RawMessage, recordsPerRequest)
	errs := make(chan error, 1)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go searchRecords(ctx, searchRequest{
		appToken:          appToken,
		records:           records,
		errs:              errs,
		timeOutSecs:       *timeOutSecs,
		limit:             limit,
		recordsPerRequest: recordsPerRequest,
		expDate:           *expDate,
		licenseNumber:     *licenseNumber,
		licenseType:       *licenseType,
		businessCounty:    *businessCounty,
		licenseSubType:    *licenseSubType,
		businessName:      *businessName,
		ownerName:         *ownerName,
	})

	count := 0
	for {
		select {
		case record, ok := <-records:
			if ok {
				count++
				fmt.Fprintf(os.Stderr, "%s\n", pretty.Color(pretty.Pretty(record), nil))
			} else {
				records = nil
			}
		case err, ok := <-errs:
			if ok {
				fmt.Fprintf(os.Stderr, "There was an error while processing your request: %v\n", err)
			} else {
				errs = nil
			}
		}

		if records == nil && errs == nil {
			break
		}
	}

	fmt.Printf("Found %d total licenses\n", count)
}

type searchRequest struct {
	appToken          string
	records           chan<- json.RawMessage
	errs              chan<- error
	timeOutSecs       int
	limit             int
	recordsPerRequest int
	expDate           string
	licenseNumber     string
	licenseType       string
	businessCounty    string
	licenseSubType    string
	businessName      string
	ownerName         string
}

func searchRecords(ctx context.Context, sReq searchRequest) {
	defer close(sReq.records)
	defer close(sReq.errs)
	client := http.Client{Timeout: time.Duration(sReq.timeOutSecs) * time.Second}
	recordsFound := 0
	offset := 0
	for {
		select {
		case <-ctx.Done():
			sReq.errs <- ctx.Err()
			return
		default:
			recordsPerRequest := sReq.recordsPerRequest
			if sReq.limit > 0 {
				remaining := sReq.limit - recordsFound
				if remaining < recordsPerRequest {
					recordsPerRequest = remaining
				}
			}

			if recordsPerRequest == 0 {
				return
			}

			whereClause := buildWhereClause(sReq)

			params := url.Values{}
			if whereClause != "" {
				params.Add("$where", whereClause)
			}

			params.Add("$limit", fmt.Sprintf("%d", recordsPerRequest))
			params.Add("$offset", fmt.Sprintf("%d", offset))

			fullURL := apiBaseURL + "?" + params.Encode()

			req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
			if err != nil {
				sReq.errs <- fmt.Errorf("error creating HTTP request: %w", err)
				return
			}

			req.Header.Add("Accept", "application/json")
			req.Header.Add("X-App-Token", sReq.appToken)

			resp, err := client.Do(req)
			if err != nil {
				sReq.errs <- fmt.Errorf("error making HTTP request: %w", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				sReq.errs <- fmt.Errorf("api returned a non-200 status code: %d %s", resp.StatusCode, resp.Status)
				return
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				sReq.errs <- fmt.Errorf("error reading response body: %w", err)
				return
			}

			var recordsRetrieved []json.RawMessage
			if err := json.Unmarshal(body, &recordsRetrieved); err != nil {
				sReq.errs <- fmt.Errorf("error unmarshaling JSON: %w", err)
				return
			}

			if len(recordsRetrieved) == 0 {
				return
			}

			for _, record := range recordsRetrieved {
				select {
				case <-ctx.Done():
					sReq.errs <- ctx.Err()
					return
				case sReq.records <- record:
				}
				recordsFound++
			}

			offset += len(recordsRetrieved)
		}
	}
}

func buildWhereClause(sReq searchRequest) string {
	var conditions []string

	if sReq.expDate != "" {
		conditions = append(conditions, fmt.Sprintf("upper(license_expiration_date_mmddccyy) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.expDate, "'", "''"))))
	}

	if sReq.licenseNumber != "" {
		conditions = append(conditions, fmt.Sprintf("upper(license_number) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.licenseNumber, "'", "''"))))
	}

	if sReq.licenseType != "" {
		conditions = append(conditions, fmt.Sprintf("upper(license_type) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.licenseType, "'", "''"))))
	}

	if sReq.businessCounty != "" {
		conditions = append(conditions, fmt.Sprintf("upper(business_county) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.businessCounty, "'", "''"))))
	}

	if sReq.licenseSubType != "" {
		conditions = append(conditions, fmt.Sprintf("upper(license_subtype) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.licenseSubType, "'", "''"))))
	}

	if sReq.businessName != "" {
		conditions = append(conditions, fmt.Sprintf("upper(business_name) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.businessName, "'", "''"))))
	}

	if sReq.ownerName != "" {
		conditions = append(conditions, fmt.Sprintf("upper(owner_name) like '%%%s%%'",
			strings.ToUpper(strings.ReplaceAll(sReq.ownerName, "'", "''"))))
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " AND ")
}
