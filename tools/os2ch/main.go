// Command os2ch copies documents from OpenSearch into ClickHouse and reports
// what did not fit.
//
// ClickHouse drops a JSON key that matches no column without saying anything,
// so a migration that "succeeded" can still have lost half of every document.
// This reads the destination's columns first and reports every key that has
// nowhere to go, per dataType — which is the only way to know whether the
// schema survives real data before committing to the whole set.
//
//	os2ch -sample 500          # a slice of every dataType, nothing written
//	os2ch -sample 500 -write   # same slice, written
//	os2ch -write               # everything
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type config struct {
	osURL        string
	osUser       string
	osPass       string
	chURL        string
	chUser       string
	chPass       string
	chDB         string
	indexGlob    string
	table        string
	sampleSize   int
	batchSize    int
	chNativePort int
	write        bool
}

func main() {
	var c config
	flag.StringVar(&c.osURL, "os", "https://localhost:9200", "OpenSearch base URL")
	flag.StringVar(&c.osUser, "os-user", "admin", "OpenSearch user")
	flag.StringVar(&c.osPass, "os-pass", "", "OpenSearch password")
	flag.StringVar(&c.chURL, "ch", "http://localhost:8123", "ClickHouse HTTP URL")
	flag.StringVar(&c.chUser, "ch-user", "default", "ClickHouse user")
	flag.StringVar(&c.chPass, "ch-pass", "", "ClickHouse password")
	flag.StringVar(&c.chDB, "ch-db", "utmstack", "ClickHouse database")
	flag.StringVar(&c.indexGlob, "index", "v11-log-*", "OpenSearch index pattern")
	flag.StringVar(&c.table, "table", "logs", "destination table")
	flag.IntVar(&c.sampleSize, "sample", 0, "stop after this many documents per dataType (0 = all)")
	flag.IntVar(&c.chNativePort, "ch-port", 9000, "ClickHouse native port")
	flag.IntVar(&c.batchSize, "batch", 2000, "documents per scroll page and per insert")
	flag.BoolVar(&c.write, "write", false, "actually write; without it nothing is inserted")
	flag.Parse()

	if c.write {
		if err := chConnect(c); err != nil {
			fail("connecting to ClickHouse", err)
		}
	}

	cols, err := columns(c)
	if err != nil {
		fail("reading the destination columns", err)
	}
	fmt.Printf("destino %s.%s: %d columnas\n", c.chDB, c.table, len(cols))
	if !c.write {
		fmt.Println("modo simulacion: no se escribe nada")
	}
	fmt.Println()

	r := &report{orphans: map[string]map[string]int{}, seen: map[string]int{}, written: map[string]int{}}
	if err := migrate(c, cols, r); err != nil {
		fail("migrating", err)
	}
	r.print()
}

// columns is what a document is measured against: anything not here is dropped
// by ClickHouse in silence.
func columns(c config) (map[string]bool, error) {
	q := fmt.Sprintf("SELECT name FROM system.columns WHERE database='%s' AND table='%s' FORMAT TSV", c.chDB, c.table)
	body, err := chQuery(c, q)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		if line != "" {
			out[line] = true
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("table %s.%s has no columns", c.chDB, c.table)
	}
	return out, nil
}

type report struct {
	// orphans maps dataType -> key -> how many documents carried it.
	orphans map[string]map[string]int
	seen    map[string]int
	written map[string]int
}

func (r *report) note(dataType string, doc map[string]json.RawMessage, cols map[string]bool) {
	r.seen[dataType]++
	for k := range doc {
		if cols[k] {
			continue
		}
		if r.orphans[dataType] == nil {
			r.orphans[dataType] = map[string]int{}
		}
		r.orphans[dataType][k]++
	}
}

func (r *report) print() {
	types := make([]string, 0, len(r.seen))
	for t := range r.seen {
		types = append(types, t)
	}
	sort.Strings(types)

	fmt.Println("dataType                        leidos  escritos  claves sin columna")
	fmt.Println(strings.Repeat("-", 78))
	totalOrphan := 0
	for _, t := range types {
		n := len(r.orphans[t])
		totalOrphan += n
		mark := ""
		if n > 0 {
			mark = "  <-- revisar"
		}
		fmt.Printf("%-28s %8d  %8d  %6d%s\n", t, r.seen[t], r.written[t], n, mark)
	}

	if totalOrphan == 0 {
		fmt.Println("\nninguna clave se quedo sin columna")
		return
	}

	fmt.Println("\nclaves que ClickHouse habria descartado en silencio:")
	for _, t := range types {
		if len(r.orphans[t]) == 0 {
			continue
		}
		keys := make([]string, 0, len(r.orphans[t]))
		for k := range r.orphans[t] {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return r.orphans[t][keys[i]] > r.orphans[t][keys[j]] })
		fmt.Printf("\n  %s\n", t)
		for i, k := range keys {
			if i >= 15 {
				fmt.Printf("      ... y %d mas\n", len(keys)-15)
				break
			}
			fmt.Printf("      %-40s en %d documentos\n", k, r.orphans[t][k])
		}
	}
}

// migrate walks the indices one at a time rather than scrolling the whole
// pattern. Log indices carry the dataType in their name, so a per-index cap is
// a per-dataType cap — and a type with five documents is read in full instead
// of being hunted for inside forty million.
func migrate(c config, cols map[string]bool, r *report) error {
	indices, err := listIndices(c)
	if err != nil {
		return err
	}
	fmt.Printf("%d indices que copiar\n\n", len(indices))

	for _, idx := range indices {
		n, w, err := copyIndex(c, idx, cols, r)
		if err != nil {
			return fmt.Errorf("%s: %w", idx, err)
		}
		fmt.Printf("  %-52s %7d leidos  %7d escritos\n", idx, n, w)
	}
	return nil
}

func listIndices(c config) ([]string, error) {
	body, err := osRequest(c, http.MethodGet, "/_cat/indices/"+c.indexGlob+"?h=index&s=index", nil)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

func copyIndex(c config, index string, cols map[string]bool, r *report) (int, int, error) {
	scrollID := ""
	defer func() {
		if scrollID != "" {
			_, _ = osRequest(c, http.MethodDelete, "/_search/scroll",
				[]byte(fmt.Sprintf(`{"scroll_id":%q}`, scrollID)))
		}
	}()

	size := c.batchSize
	if c.sampleSize > 0 && c.sampleSize < size {
		size = c.sampleSize
	}

	body, err := osRequest(c, http.MethodPost,
		fmt.Sprintf("/%s/_search?scroll=2m&size=%d", index, size),
		[]byte(`{"query":{"match_all":{}},"sort":["_doc"]}`))
	if err != nil {
		return 0, 0, err
	}

	read, written := 0, 0
	for {
		var page struct {
			ScrollID string `json:"_scroll_id"`
			Hits     struct {
				Hits []struct {
					Source map[string]json.RawMessage `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return read, written, fmt.Errorf("decoding a page: %w", err)
		}
		scrollID = page.ScrollID
		if len(page.Hits.Hits) == 0 {
			return read, written, nil
		}

		batch := make([][]byte, 0, len(page.Hits.Hits))
		for _, h := range page.Hits.Hits {
			if c.sampleSize > 0 && read >= c.sampleSize {
				break
			}
			dataType := jsonString(h.Source["dataType"])
			if dataType == "" {
				dataType = "(sin dataType)"
			}

			read++
			r.note(dataType, h.Source, cols)

			if c.write {
				line, err := json.Marshal(h.Source)
				if err != nil {
					continue
				}
				batch = append(batch, line)
				written++
				r.written[dataType]++
			}
		}

		if len(batch) > 0 {
			if err := chInsert(c, batch); err != nil {
				return read, written, fmt.Errorf("inserting: %w", err)
			}
		}

		if c.sampleSize > 0 && read >= c.sampleSize {
			return read, written, nil
		}

		body, err = osRequest(c, http.MethodPost, "/_search/scroll",
			[]byte(fmt.Sprintf(`{"scroll":"2m","scroll_id":%q}`, scrollID)))
		if err != nil {
			return read, written, err
		}
	}
}

var chConn driver.Conn

func chConnect(c config) error {
	host := strings.TrimPrefix(strings.TrimPrefix(c.chURL, "http://"), "https://")
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, c.chNativePort)},
		Auth: clickhouse.Auth{Database: c.chDB, Username: c.chUser, Password: c.chPass},
	})
	if err != nil {
		return err
	}
	chConn = conn
	return conn.Ping(context.Background())
}

// chInsert writes over the native protocol rather than HTTP. The HTTP
// interface rejects both the RFC3339 timestamps and the nested objects these
// documents carry; the plugins use the native path, so using it here means a
// document that lands during migration is one that would land in production.
func chInsert(c config, rows [][]byte) error {
	batch := fmt.Sprintf("INSERT INTO %s.%s FORMAT JSONEachRow\n%s",
		c.chDB, c.table, bytes.Join(rows, []byte("\n")))
	return chConn.AsyncInsert(context.Background(), batch, true)
}

func chQuery(c config, q string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/?user=%s&password=%s", c.chURL, urlEscape(c.chUser), urlEscape(c.chPass)),
		strings.NewReader(q))
	if err != nil {
		return nil, err
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clickhouse said %d: %s", resp.StatusCode, truncate(string(b)))
	}
	return b, nil
}

func osRequest(c config, method, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, c.osURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.osUser, c.osPass)

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("opensearch said %d: %s", resp.StatusCode, truncate(string(b)))
	}
	return b, nil
}

func httpClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			// The dev cluster presents a self-signed certificate.
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func jsonString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return s
}

func urlEscape(s string) string {
	r := strings.NewReplacer(" ", "%20", "'", "%27", "=", "%3D", "&", "%26",
		"*", "%2A", "?", "%3F", "+", "%2B", "\n", "%0A")
	return r.Replace(s)
}

func truncate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}

func fail(what string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", what, err)
	os.Exit(1)
}
