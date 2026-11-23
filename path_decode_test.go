package flow

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPathParameterDecoding(t *testing.T) {
	mux := New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Query-Value", r.URL.Query().Get("q"))
		io.WriteString(w, r.PathValue("value"))
	})
	mux.Handle("/plain/:value", handler, http.MethodGet)
	mux.Handle("/regexp/:value|^a[+]b$", handler, http.MethodGet)
	server := httptest.NewServer(mux)
	defer server.Close()

	cases := []struct {
		path string
		want string
	}{
		{"/plain/a+b", "a+b"},
		{"/plain/a%2Bb", "a+b"},
		{"/plain/a%20b", "a b"},
		{"/plain/a%252Bb", "a%2Bb"},
		{"/plain/a%2Fb", "a/b"},
		{"/plain/%E2%9C%93+ok", "✓+ok"},
		{"/regexp/a+b", "a+b"},
		{"/regexp/a%2Bb", "a+b"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			res, err := server.Client().Get(server.URL + tc.path + "?q=a+b")
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != http.StatusOK || string(body) != tc.want {
				t.Fatalf("got %d %q; want 200 %q", res.StatusCode, body, tc.want)
			}
			if got := res.Header.Get("X-Query-Value"); got != "a b" {
				t.Fatalf("query decoding changed: got %q; want %q", got, "a b")
			}
		})
	}
}
