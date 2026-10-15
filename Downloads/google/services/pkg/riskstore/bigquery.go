package riskstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/finguard/platform/pkg/domain"
)

// BigQueryClient reads customer transactions from a BigQuery-compatible
// HTTP endpoint. The goccy BigQuery emulator speaks the BigQuery REST
// query API and is auth-less, so we don't need the full cloud.google.com/go/bigquery
// SDK or any credentials to talk to it. The same code path will work
// against real BigQuery with the addition of an Authorization header.
//
// Configuration is via Host (e.g. "http://127.0.0.1:9050"), Project
// (e.g. "finguard-dev") and Dataset (e.g. "finguard_dev"). The dataset
// must already contain a "transactions" table matching the BigQuery DDL
// in terraform/modules/data/schemas/transactions.sql.
//
// Returned envelopes have Simulated=false and Source="riskstore:BigQueryClient".
type BigQueryClient struct {
	Host    string
	Project string
	Dataset string
	// HTTPClient allows tests to inject a custom client. Defaults to a
	// 3-second timeout http.Client.
	HTTPClient *http.Client
}

// bqQueryResponse is the subset of the BigQuery REST query response we
// consume. Only the fields we read are typed; the rest is ignored.
type bqQueryResponse struct {
	Schema struct {
		Fields []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"fields"`
	} `json:"schema"`
	Rows []struct {
		F []struct {
			V string `json:"v"`
		} `json:"f"`
	} `json:"rows"`
	TotalRows string `json:"totalRows"`
	JobComplete bool   `json:"jobComplete"`
}

// ListCustomerTransactions returns the customer's most-recent transactions
// from BigQuery, ordered by event_time DESC, capped by limit (0 -> 50).
func (c *BigQueryClient) ListCustomerTransactions(ctx context.Context, customerID string, limit int) (domain.CustomerTransactions, error) {
	if c.Host == "" {
		return domain.CustomerTransactions{}, fmt.Errorf("riskstore: BigQueryClient.Host is empty")
	}
	if c.Project == "" || c.Dataset == "" {
		return domain.CustomerTransactions{}, fmt.Errorf("riskstore: BigQueryClient.Project/Dataset must be set")
	}
	if limit <= 0 {
		limit = 50
	}
	// The goccy emulator does not support query parameters, so we inline
	// the customer id with single-quote escaping. The BFF already restricts
	// the id to /^[A-Za-z0-9._-]{1,64}$/ (defense in depth: don't trust
	// the upstream regex; reject anything that could break the SQL).
	if !isSafeBQString(customerID) {
		return domain.CustomerTransactions{}, fmt.Errorf("riskstore: unsafe customer id")
	}

	q := fmt.Sprintf(
		"SELECT txn_id, customer_id, amount, currency, merchant_id, merchant_name, mcc, device_id, ip_address, geo_country, event_time, source FROM `%s.%s.transactions` WHERE customer_id = '%s' ORDER BY event_time DESC LIMIT %d",
		c.Project, c.Dataset, customerID, limit,
	)

	body, err := c.query(ctx, q)
	if err != nil {
		return domain.CustomerTransactions{}, err
	}
	if !body.JobComplete {
		return domain.CustomerTransactions{}, fmt.Errorf("riskstore: query not complete")
	}

	// Build a name -> position index so we don't depend on column order.
	idx := make(map[string]int, len(body.Schema.Fields))
	for i, f := range body.Schema.Fields {
		idx[f.Name] = i
	}
	must := func(name string) int {
		i, ok := idx[name]
		if !ok {
			return -1
		}
		return i
	}
	txnIdx := must("txn_id")
	cidIdx := must("customer_id")
	amtIdx := must("amount")
	curIdx := must("currency")
	midIdx := must("merchant_id")
	mnameIdx := must("merchant_name")
	mccIdx := must("mcc")
	devIdx := must("device_id")
	ipIdx := must("ip_address")
	geoIdx := must("geo_country")
	tIdx := must("event_time")
	srcIdx := must("source")

	out := make([]domain.Transaction, 0, len(body.Rows))
	for _, row := range body.Rows {
		if txnIdx < 0 || txnIdx >= len(row.F) {
			continue
		}
		tx := domain.Transaction{
			TxnID:      cell(row, txnIdx),
			CustomerID: cell(row, cidIdx),
			Currency:   cell(row, curIdx),
			MerchantID: cell(row, midIdx),
			DeviceID:   cell(row, devIdx),
			IPAddress:  cell(row, ipIdx),
			Simulated:  false,
		}
		// merchant_category derived from MCC code is not done here; pass
		// the merchant_name through into the merchant_category field for
		// display purposes so the UI has something to show.
		if mnameIdx >= 0 {
			tx.MerchantCat = cell(row, mnameIdx)
		}
		// If the schema has an explicit mcc but no merchant_name, expose
		// the MCC code as a synthetic category.
		if tx.MerchantCat == "" && mccIdx >= 0 {
			tx.MerchantCat = cell(row, mccIdx)
		}
		if amtIdx >= 0 {
			if f, err := strconv.ParseFloat(cell(row, amtIdx), 64); err == nil {
				tx.Amount = f
			}
		}
		if geoIdx >= 0 {
			tx.Geo = domain.Geo{Lat: 0, Lon: 0} // BQ has only country in our schema; lat/lon default to 0
		}
		if srcIdx >= 0 {
			tx.Channel = domain.Channel(cell(row, srcIdx))
		}
		if tIdx >= 0 {
			if t, err := time.Parse(time.RFC3339, cell(row, tIdx)); err == nil {
				tx.Timestamp = t
			} else {
				tx.Timestamp = time.Now().UTC()
			}
		} else {
			tx.Timestamp = time.Now().UTC()
		}
		out = append(out, tx)
	}

	return domain.CustomerTransactions{
		CustomerID:   customerID,
		Count:        len(out),
		Simulated:    false,
		Source:       "riskstore:BigQueryClient",
		Transactions: out,
	}, nil
}

// query POSTs the SQL to the BigQuery REST query endpoint and decodes
// the JSON response. The goccy emulator doesn't support useLegacySql
// or parameterMode, so we send a bare {"query": "..."} body.
func (c *BigQueryClient) query(ctx context.Context, q string) (*bqQueryResponse, error) {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 3 * time.Second}
	}
	endpoint := fmt.Sprintf("%s/bigquery/v2/projects/%s/queries",
		strings.TrimRight(c.Host, "/"),
		url.PathEscape(c.Project),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader([]byte(`{"query":`+jsonString(q)+`}`)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("riskstore: bq %s: %s", res.Status, string(raw))
	}
	var out bqQueryResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("riskstore: bq decode: %w (body=%s)", err, string(raw))
	}
	return &out, nil
}

func cell(row struct {
	F []struct {
		V string `json:"v"`
	} `json:"f"`
}, i int) string {
	if i < 0 || i >= len(row.F) {
		return ""
	}
	return row.F[i].V
}

// jsonString encodes a string as a JSON string literal, escaping control
// characters and the bytes that would otherwise break out of the literal.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// isSafeBQString reports whether s contains only characters that are
// safe to inline into a single-quoted SQL literal. We disallow quotes,
// backslashes, and newlines, which are the only bytes that matter for
// an injection.
func isSafeBQString(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z',
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.':
		default:
			return false
		}
	}
	return true
}
