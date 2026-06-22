package fbhttp

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

type modifyRequest struct {
	What            string   `json:"what"`
	Which           []string `json:"which"`
	CurrentPassword string   `json:"current_password"`
}

func NewHandler(
	imgSvc ImgService,
	fileCache FileCache,
	uploadCache UploadCache,
	store *storage.Storage,
	server *settings.Server,
	assetsFs fs.FS,
) (http.Handler, error) {
	server.Clean()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
			next.ServeHTTP(w, r)
		})
	})
	index, static := getStaticHandlers(store, server, assetsFs)

	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	r.HandleFunc("/health", healthHandler)
	r.PathPrefix("/static").Handler(static)
	r.PathPrefix("/webdav").Handler(monkey(webDavHandler, ""))
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	tokenExpirationTime := server.GetTokenExpirationTime(DefaultTokenExpiration)
	authHandler := &auth.Handler{
		ExtractToken: auth.ExtractToken,
		RevokeToken:  auth.RevokeToken,
		RevokeAll:    auth.RevokeAllUserTokens,
		KeyFunc: func() (interface{}, error) {
			return server.Key, nil
		},
		TokenExpirationTime: tokenExpirationTime,
		Method:              server.AuthMethod,
		Hook:                server.AuthHook,
	}

	api.Use(authHandler.Middleware)

	api.HandleFunc("/login", loginHandler).Methods("GET")
	api.Path("/login").HandlerFunc(makeAuthHandler(auth.MethodJSONAuth)).Methods("POST")
	api.Path("/signup").HandlerFunc(signupHandler).Methods("POST")
	api.Path("/renew").HandlerFunc(renewHandler).Methods("GET")

	api.Path("/settings").Handler(monkey(settingsGetHandler, "")).Methods("GET")
	api.Path("/settings").Handler(monkey(settingsPutHandler, "")).Methods("PUT")

	api.Path("/users").Handler(monkey(usersGetHandler, "")).Methods("GET")
	api.Path("/users/{id}" ).Handler(monkey(userGetHandler, "")).Methods("GET")
	api.Path("/users" ).Handler(monkey(userPostHandler, "")).Methods("POST")
	api.Path("/users/{id}" ).Handler(monkey(userPutHandler, "")).Methods("PUT")
	api.Path("/users/{id}" ).Handler(monkey(userDeleteHandler, "")).Methods("DELETE")

	api.PathPrefix("/files").Handler(monkey(filesHandler, "/api/files")).Methods("GET")
	api.PathPrefix("/files").Handler(monkey(resourcePostHandler, "/api/files")).Methods("POST")
	api.PathPrefix("/files").Handler(monkey(resourcePutHandler, "/api/files")).Methods("PUT")
	api.PathPrefix("/files").Handler(monkey(resourcePatchHandler, "/api/files")).Methods("PATCH")
	api.PathPrefix("/files").Handler(monkey(resourceDeleteHandler, "/api/files")).Methods("DELETE")

	api.PathPrefix("/tus").Handler(monkey(tusHandler, "/api/tus")).Methods("OPTIONS")
	api.PathPrefix("/tus").Handler(monkey(tusHandler, "/api/tus")).Methods("POST", "HEAD", "PATCH", "DELETE")

	api.Path("/resources").Handler(monkey(resourceGetHandler, "")).Methods("GET")
	api.PathPrefix("/preview").Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	api.PathPrefix("/download").Handler(monkey(downloadHandler, "/api/download")).Methods("GET")

	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")
	api.PathPrefix("/resources").Handler(monkey(resourcesHandler, "/api/resources")).Methods("GET")

	api.Path("/usage").Handler(monkey(usageHandler, "")).Methods("GET")

	api.Path("/shares").Handler(monkey(shareListHandler, "")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(shareGetsHandler, "/api/share")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(sharePostHandler, "/api/share")).Methods("POST")
	api.PathPrefix("/share").Handler(monkey(shareDeleteHandler, "/api/share")).Methods("DELETE")

	api.Handle("/settings", monkey(settingsGetHandler, "")).Methods("GET")
	api.Handle("/settings", monkey(settingsPutHandler, "")).Methods("PUT")

	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	api.PathPrefix("/preview/{size}/{path:.*}").
		Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	api.PathPrefix("/command").Handler(monkey(commandsHandler, "/api/command")).Methods("GET")
	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")
	api.PathPrefix("/subtitle").Handler(monkey(subtitleHandler, "/api/subtitle")).Methods("GET")

	public := api.PathPrefix("/public").Subrouter()
	public.PathPrefix("/dl").Handler(monkey(publicDlHandler, "/api/public/dl/")).Methods("GET")
	public.PathPrefix("/share").Handler(monkey(publicShareHandler, "/api/public/share/")).Methods("GET")

	return stripPrefix(server.BaseURL, r), nil
}
