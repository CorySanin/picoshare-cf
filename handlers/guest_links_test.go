package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mtlynch/picoshare/v2/handlers"
	"github.com/mtlynch/picoshare/v2/picoshare"
	"github.com/mtlynch/picoshare/v2/store"
	"github.com/mtlynch/picoshare/v2/store/test_sqlite"
)

func TestGuestLinksPostAcceptsValidRequest(t *testing.T) {
	for _, tt := range []struct {
		description string
		payload     string
		expected    picoshare.GuestLink
	}{
		{
			description: "minimally populated request",
			payload: `{
					"label": null,
					"urlExpirationTime":"2030-01-02T03:04:25Z",
					"fileLifetime":"876000h0m0s",
					"maxFileBytes": null,
					"maxFileUploads": null
				}`,
			expected: picoshare.GuestLink{
				Label:          picoshare.GuestLinkLabel(""),
				UrlExpires:     mustParseExpirationTime("2030-01-02T03:04:25Z"),
				FileLifetime:   picoshare.FileLifetimeInfinite,
				MaxFileBytes:   picoshare.GuestUploadUnlimitedFileSize,
				MaxFileUploads: picoshare.GuestUploadUnlimitedFileUploads,
			},
		},
		{
			description: "fully populated request",
			payload: `{
					"label": "For my good pal, Maurice",
					"urlExpirationTime":"2030-01-02T03:04:25Z",
					"fileLifetime":"876000h0m0s",
					"maxFileBytes": 1048576,
					"maxFileUploads": 1
				}`,
			expected: picoshare.GuestLink{
				Label:          picoshare.GuestLinkLabel("For my good pal, Maurice"),
				UrlExpires:     mustParseExpirationTime("2030-01-02T03:04:25Z"),
				FileLifetime:   picoshare.FileLifetimeInfinite,
				MaxFileBytes:   makeGuestUploadMaxFileBytes(1048576),
				MaxFileUploads: makeGuestUploadCountLimit(1),
			},
		},
		{
			description: "guest file expires in 1 day",
			payload: `{
					"label": "For my good pal, Maurice",
					"urlExpirationTime":"2030-01-02T03:04:25Z",
					"fileLifetime":"24h0m0s",
					"maxFileBytes": 1048576,
					"maxFileUploads": 1
				}`,
			expected: picoshare.GuestLink{
				Label:          picoshare.GuestLinkLabel("For my good pal, Maurice"),
				UrlExpires:     mustParseExpirationTime("2030-01-02T03:04:25Z"),
				FileLifetime:   picoshare.NewFileLifetimeInDays(1),
				MaxFileBytes:   makeGuestUploadMaxFileBytes(1048576),
				MaxFileUploads: makeGuestUploadCountLimit(1),
			},
		},
		{
			description: "guest file expires in 30 day",
			payload: `{
					"label": "For my good pal, Maurice",
					"urlExpirationTime":"2030-01-02T03:04:25Z",
					"fileLifetime":"720h0m0s",
					"maxFileBytes": 1048576,
					"maxFileUploads": 1
				}`,
			expected: picoshare.GuestLink{
				Label:          picoshare.GuestLinkLabel("For my good pal, Maurice"),
				UrlExpires:     mustParseExpirationTime("2030-01-02T03:04:25Z"),
				FileLifetime:   picoshare.NewFileLifetimeInDays(30),
				MaxFileBytes:   makeGuestUploadMaxFileBytes(1048576),
				MaxFileUploads: makeGuestUploadCountLimit(1),
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()
			s := handlers.New(mockAuthenticator{}, &dataStore, nilSpaceChecker, nilGarbageCollector, handlers.NewClock())

			req, err := http.NewRequest("POST", "/api/guest-links", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if status := res.StatusCode; status != http.StatusOK {
				t.Fatalf("%s: handler returned wrong status code: got %v want %v",
					tt.description, status, http.StatusOK)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body")
			}

			var response handlers.GuestLinkPostResponse
			err = json.Unmarshal(body, &response)
			if err != nil {
				t.Fatalf("response is not valid JSON: %v", body)
			}

			gl, err := dataStore.GetGuestLink(picoshare.GuestLinkID(response.ID))
			if err != nil {
				t.Fatalf("%s: failed to retrieve guest link from datastore: %v", tt.description, err)
			}

			// Copy the values that we can't predict in advance.
			tt.expected.ID = picoshare.GuestLinkID(response.ID)
			tt.expected.Created = gl.Created

			if !reflect.DeepEqual(gl, tt.expected) {
				t.Fatalf("%s: guest link does not match expected: got %+v, want %+v", tt.description, gl, tt.expected)
			}
		})
	}
}

func TestGuestLinksPostRejectsInvalidRequest(t *testing.T) {
	for _, tt := range []struct {
		description string
		payload     string
	}{
		{
			description: "empty string",
			payload:     "",
		},
		{
			description: "empty payload",
			payload:     "{}",
		},
		{
			description: "invalid label field (non-string)",
			payload: `{
					"label": 5,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": null,
					"maxFileUploads": null
				}`,
		},
		{
			description: "invalid label field (too long)",
			payload: fmt.Sprintf(`{
					"label": "%s",
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": null,
					"maxFileUploads": null
				}`, strings.Repeat("A", 201)),
		},
		{
			description: "missing urlExpirationTime field",
			payload: `{
					"label": null,
					"maxFileBytes": null,
					"maxFileUploads": null
				}`,
		},
		{
			description: "invalid expirationTime field",
			payload: `{
					"label": null,
					"urlExpirationTime": 25,
					"maxFileBytes": null,
					"maxFileUploads": null
				}`,
		},
		{
			description: "negative maxFileBytes field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": -5,
					"maxFileUploads": null
				}`,
		},
		{
			description: "decimal maxFileBytes field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": 1.5,
					"maxFileUploads": null
				}`,
		},
		{
			description: "too low a maxFileBytes field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": 1,
					"maxFileUploads": null
				}`,
		},
		{
			description: "zero maxFileBytes field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": 0,
					"maxFileUploads": null
				}`,
		},
		{
			description: "negative maxFileUploads field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": null,
					"maxFileUploads": -5
				}`,
		},
		{
			description: "decimal maxFileUploads field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": null,
					"maxFileUploads": 1.5
				}`,
		},
		{
			description: "zero maxFileUploads field",
			payload: `{
					"label": null,
					"urlExpirationTime":"2025-01-01T00:00:00Z",
					"maxFileBytes": null,
					"maxFileUploads": 0
				}`,
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			dataStore := test_sqlite.New()
			s := handlers.New(mockAuthenticator{}, &dataStore, nilSpaceChecker, nilGarbageCollector, handlers.NewClock())

			req, err := http.NewRequest("POST", "/api/guest-links", strings.NewReader(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Content-Type", "text/json")

			rec := httptest.NewRecorder()
			s.Router().ServeHTTP(rec, req)
			res := rec.Result()

			if status := res.StatusCode; status != http.StatusBadRequest {
				t.Fatalf("%s: handler returned wrong status code: got %v want %v",
					tt.description, status, http.StatusBadRequest)
			}
		})
	}
}

func makeGuestUploadMaxFileBytes(i uint64) picoshare.GuestUploadMaxFileBytes {
	return picoshare.GuestUploadMaxFileBytes(&i)
}

func makeGuestUploadCountLimit(i int) picoshare.GuestUploadCountLimit {
	return picoshare.GuestUploadCountLimit(&i)
}

func TestDeleteExistingGuestLink(t *testing.T) {
	dataStore := test_sqlite.New()
	dataStore.InsertGuestLink(picoshare.GuestLink{
		ID:         picoshare.GuestLinkID("abcdefgh23456789"),
		Created:    time.Now(),
		UrlExpires: mustParseExpirationTime("2030-01-02T03:04:25Z"),
	})

	s := handlers.New(mockAuthenticator{}, &dataStore, nilSpaceChecker, nilGarbageCollector, handlers.NewClock())

	req, err := http.NewRequest("DELETE", "/api/guest-links/abcdefgh23456789", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	res := rec.Result()

	if status := res.StatusCode; status != http.StatusOK {
		t.Fatalf("DELETE returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	_, err = dataStore.GetGuestLink(picoshare.GuestLinkID("dummy-guest-link-id"))
	if _, ok := err.(store.GuestLinkNotFoundError); !ok {
		t.Fatalf("expected entry %v to be deleted, got: %v", picoshare.EntryID("abcdefgh23456789"), err)
	}
}

func TestDeleteNonExistentGuestLink(t *testing.T) {
	dataStore := test_sqlite.New()
	s := handlers.New(mockAuthenticator{}, &dataStore, nilSpaceChecker, nilGarbageCollector, handlers.NewClock())

	req, err := http.NewRequest("DELETE", "/api/guest-links/abcdefgh23456789", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	res := rec.Result()

	// File doesn't exist, but there's no error for deleting a non-existent file.
	if status := res.StatusCode; status != http.StatusOK {
		t.Fatalf("DELETE returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestDeleteInvalidGuestLink(t *testing.T) {
	dataStore := test_sqlite.New()
	s := handlers.New(mockAuthenticator{}, &dataStore, nilSpaceChecker, nilGarbageCollector, handlers.NewClock())

	req, err := http.NewRequest("DELETE", "/api/guest-links/i-am-an-invalid-link", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	res := rec.Result()

	if status := res.StatusCode; status != http.StatusBadRequest {
		t.Fatalf("DELETE returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
