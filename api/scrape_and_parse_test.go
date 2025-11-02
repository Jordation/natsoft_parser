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
	tests := []*struct {
		pathExt string
	}{
		{"0"}, {"1"}, {"2"},
		{"3"}, {"4"}, {"5"},
	}

	srv := setupTestSrv(t, "testdata/results", "result")

	time.Sleep(time.Second / 100)

	parser := NewParser()

	for _, test := range tests {
		t.Run("test for file: "+test.pathExt, func(t *testing.T) {
			entries, _, err := parser.Parse(srv.URL + urlSplitter + test.pathExt + ".html")
			if err != nil {
				t.Fatalf("failed to translate page:%s", err)
			}

			if err := parser.WriteEntriesTo(entries, "testdata/results/output_"+test.pathExt+".csv"); err != nil {
				t.Fatalf("failed to write result csv:%s", err)
			}
		})
	}

}

func TestTimesParse(t *testing.T) {
	const urlSplitter = "/*"
	tests := []*struct {
		pathExt string
	}{
		{"0"}, {"1"}, {"2"},
		{"3"}, {"4"}, {"5"},
	}

	srv := setupTestSrv(t, "testdata/lap_times", "times")

	time.Sleep(time.Second / 100)

	parser := NewParser()

	for _, test := range tests {
		t.Run("test for file: "+test.pathExt, func(t *testing.T) {
			entries, _, err := parser.Parse(srv.URL + urlSplitter + test.pathExt + ".html")
			if err != nil {
				t.Fatalf("failed to translate page:%s", err)
			}

			if err := parser.WriteEntriesTo(entries, "testdata/lap_times/output_"+test.pathExt+".csv"); err != nil {
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
