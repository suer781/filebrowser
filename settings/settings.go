package settings

import (
	"crypto/rand"
	"io/fs"
	"log"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/rules"
)

const DefaultUsersHomeBasePath = "/users"
const DefaultLogoutPage = "/login"
const DefaultMinimumPasswordLength = 12
const DefaultFileMode = 0640
const DefaultDirMode = 0750

type AuthMethod string

type Settings struct {
	Key                   []byte              `json:"key"`
	Signup                bool                `json:"signup"`
	HideLoginButton       bool                `json:"hideLoginButton"`
	CreateUserDir         bool                `json:"createUserDir"`
	UserHomeBasePath      string              `json:"userHomeBasePath"`
	Defaults              UserDefaults        `json:"defaults"`
	AuthMethod            AuthMethod          `json:"authMethod"`
	LogoutPage            string              `json:"logoutPage"`
	Branding              Branding            `json:"branding"`
	Tus                   Tus                 `json:"tus"`
	Commands              map[string][]string `json:"commands"`
	Shell                 []string            `json:"shell"`
	Rules                 []rules.Rule        `json:"rules"`
	MinimumPasswordLength uint                `json:"minimumPasswordLength"`
	FileMode              fs.FileMode         `json:"fileMode"`
	DirMode               fs.FileMode         `json:"dirMode"`
	EnableWebDAV          bool                `json:"enableWebDAV"`
}

type Branding struct {
	Name        string `json:"name"`
	DisableExternal bool `json:"disableExternal"`
}

type UserDefaults struct {
	Mode         uint8  `json:"mode"`
	Locale       string `json:"locale"`
	ShowHidden    bool   `json:"showHidden"`
	SingleClick  bool   `json:"singleClick"`
	Sorting      string `json:"sorting"`
	Perm         UserPermissions `json:"perm"`
}

type UserPermissions struct {
	Admin    bool `json:"admin"`
	Execute  bool `json:"execute"`
	Create   bool `json:"create"`
	Rename   bool `json:"rename"`
	Modify   bool `json:"modify"`
	Delete   bool `json:"delete"`
	Share    bool `json:"share"`
	Download bool `json:"download"`
}

type Tus struct {
	ChunkSize int64  `json:"chunkSize"`
	Resumable bool   `json:"resumable"`
	DataDir   string `json:"dataDir"`
}

type Server struct {
	BaseURL              string `json:"baseURL"`
	Socket               string `json:"socket"`
	TLSKey               string `json:"tlsKey"`
	TLSCert              string `json:"tlsCert"`
	Port                 string `json:"port"`
	Address              string `json:"address"`
	Log                  string `json:"log"`
	EnableThumbnails     bool   `json:"enableThumbnails"`
	ResizePreview        bool   `json:"resizePreview"`
	EnableExec           bool   `json:"enableExec"`
	TypeDetectionByHeader bool  `json:"typeDetectionByHeader"`
	ImageResolutionCal   bool   `json:"imageResolutionCalculation"`
	AuthHook             string `json:"authHook"`
	TokenExpirationTime  string `json:"tokenExpirationTime"`
}

func (s *Server) Clean() {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")
}

func (s *Server) GetTokenExpirationTime(fallback time.Duration) time.Duration {
	if s.TokenExpirationTime == "" {
		return fallback
	}
	duration, err := time.ParseDuration(s.TokenExpirationTime)
	if err != nil {
		log.Printf("[WARN] Failed to parse tokenExpirationTime: %v", err)
		return fallback
	}
	return duration
}

func GenerateKey() ([]byte, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
