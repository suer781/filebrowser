package fbhttp

import (
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestWebDavHandler(t *testing.T) {
	const password = "password"
	hashedPassword, _ := users.HashPwd(password)

	testCases := map[string]struct {
		enableWebDAV       bool
		method             string
		username           string
		password           string
		expectedStatusCode int
	}{
		"WebDAV disabled": {
			enableWebDAV:       false,
			username:           "admin",
			password:           password,
			expectedStatusCode: 403,
		},
		"WebDAV enabled, no auth": {
			enableWebDAV:       true,
			username:           "",
			password:           "",
			expectedStatusCode: 401,
		},
		"WebDAV enabled, wrong password": {
			enableWebDAV:       true,
			username:           "admin",
			password:           "wrong",
			expectedStatusCode: 401,
		},
		"WebDAV enabled, correct auth": {
			enableWebDAV:       true,
			username:           "admin",
			password:           password,
			expectedStatusCode: 207,
		},
		"WebDAV enabled, correct auth, DELETE without modify": {
			enableWebDAV:       true,
			method:             "DELETE",
			username:           "admin",
			password:           password,
			expectedStatusCode: 403,
		},
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := storm.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatal(err)
	}

	hashed, _ := users.HashPwd(password)
	admin := &users.User{
		Username: "admin",
		Password: hashed,
		Perm:     users.Permissions{Modify: false},
	}
	if err := store.Users.Save(admin, admin.Password); err != nil {
		t.Fatal(err)
	}

	server := &settings.Server{BaseURL: ""}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s := &settings.Settings{
				AuthMethod:    auth.MethodJSONAuth,
				EnableWebDAV:  tc.enableWebDAV,
			}
			if err := store.Settings.Save(s); err != nil {
				t.Fatalf("failed to save settings: %v", err)
			}

			method := "PROPFIND"
			if tc.method != "" {
				method = tc.method
			}
			req := httptest.NewRequest(method, "/webdav/", nil)
			if tc.username != "" {
				req.SetBasicAuth(tc.username, tc.password)
			}

			recorder := httptest.NewRecorder()

			d := &data{
				server:   server,
				store:    store,
				settings: s,
			}

			status, err := webDavHandler(recorder, req, d)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if status != 0 {
				if status != tc.expectedStatusCode {
					t.Errorf("expected status %d, got %d", tc.expectedStatusCode, status)
				}
			} else {
				res := recorder.Result()
				if res.StatusCode != tc.expectedStatusCode {
					t.Errorf("expected status %d, got %d", tc.expectedStatusCode, res.StatusCode)
				}
			}
		})
	}
}
