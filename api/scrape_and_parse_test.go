package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestResultParse(t *testing.T) {
	const urlSplitter = "/*"
	const testType = "Results"
	tests := []*struct {
		pathExt string
	}{
		{"0"}, {"1"}, {"2"},
		{"3"}, {"4"}, {"5"},
	}

	srv := setupTestSrv(t, "testdata/results", testType)

	time.Sleep(time.Second / 100)

	parser := NewParser()

	for _, test := range tests {
		t.Run("test for file: "+test.pathExt, func(t *testing.T) {
			entries, _, err := parser.Parse(srv.URL + "/" + testType + urlSplitter + test.pathExt + ".html")
			if err != nil {
				t.Fatalf("failed to translate page:%s", err)
			}

			fileName := "testdata/results/output_" + test.pathExt + ".csv"
			file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				t.Fatal("failed opening file:", "err", err.Error(), "filename", fileName)
			}
			defer file.Close()

			if err := file.Truncate(0); err != nil {
				t.Fatal("failed to truncate fail", "err", err.Error(), "filename", fileName)
			}

			if err := parser.WriteEntriesTo(entries, file); err != nil {
				t.Fatalf("failed to write result csv:%s", err)
			}
		})
	}

}

func TestTimesParse(t *testing.T) {
	const urlSplitter = "/*"
	const testType = "Times"
	tests := []*struct {
		pathExt string
	}{
		{"0"}, {"1"}, {"2"},
		{"3"}, {"4"}, {"5"},
	}

	srv := setupTestSrv(t, "testdata/lap_times", testType)

	time.Sleep(time.Second / 100)

	parser := NewParser()

	for _, test := range tests {
		t.Run("test for file: "+test.pathExt, func(t *testing.T) {
			entries, _, err := parser.Parse(srv.URL + "/" + testType + urlSplitter + test.pathExt + ".html")
			if err != nil {
				t.Fatalf("failed to translate page:%s", err)
			}

			fileName := "testdata/results/output_" + test.pathExt + ".csv"
			file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				t.Fatal("failed opening file:", "err", err.Error(), "filename", fileName)
			}
			defer file.Close()

			if err := file.Truncate(0); err != nil {
				t.Fatal("failed to truncate fail", "err", err.Error(), "filename", fileName)
			}

			if err := parser.WriteEntriesTo(entries, file); err != nil {
				t.Fatalf("failed to write result csv:%s", err)
			}
		})
	}

}

func setupTestSrv(t *testing.T, testDir, testType string) *httptest.Server {
	return httptest.NewServer(http.Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			segs := strings.Split(r.URL.String(), "/*")
			if len(segs) != 2 {
				t.Fatalf("wrong URL format:%v", segs)
			}

			path := testDir + "/" + testType + "_" + segs[1]

			bytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to load html:%s", err)
			}

			if _, err := w.Write(bytes); err != nil {
				t.Fatalf("failed to write page:%s", err)
			}
		})),
	)
}
