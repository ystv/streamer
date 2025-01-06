package views

import (
	"encoding/hex"
	"encoding/xml"
	//nolint:gosec
	"math/rand"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/patrickmn/go-cache"

	"github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/server/store"
	"github.com/ystv/streamer/server/templates"
)

type (
	// Config the global web-auth configuration
	Config struct {
		Verbose               bool
		Version               string
		StreamServer          string `envconfig:"STREAM_SERVER"`
		TransmissionLight     string `envconfig:"TRANSMISSION_LIGHT"`
		KeyChecker            string `envconfig:"KEY_CHECKER"`
		AuthEndpoint          string `envconfig:"AUTH_ENDPOINT"`
		ServerAddress         string `envconfig:"SERVER_ADDRESS"`
		StreamerWebsocketPath string `envconfig:"STREAMER_WEBSOCKET_PATH"`
		StreamerWebAddress    string `envconfig:"STREAMER_WEB_ADDRESS"`
		StreamerAdminPath     string `envconfig:"STREAMER_ADMIN_PATH"`
		AuthenticationKey     string `envconfig:"AUTHENTICATION_KEY"`
		EncryptionKey         string `envconfig:"ENCRYPTION_KEY"`
		SessionCookieName     string `envconfig:"SESSION_COOKIE_NAME"`
	}

	// Views encapsulates our view dependencies
	Views struct {
		cache    *cache.Cache
		cookie   *sessions.CookieStore
		conf     Config
		store    *store.Store
		template *templates.Templater
	}

	// RTMP struct is the parent rtmp xml body
	RTMP struct {
		XMLName xml.Name `xml:"rtmp"`
		Server  Server   `xml:"server"`
	}

	// Server holds all the applications that will accept streams
	Server struct {
		XMLName      xml.Name      `xml:"server"`
		Applications []Application `xml:"application"`
	}

	// Application is the endpoint section that streams can go to
	Application struct {
		XMLName xml.Name `xml:"application"`
		Name    string   `xml:"name"`
		Live    Live     `xml:"live"`
	}

	// Live holds all the stream elements
	Live struct {
		XMLName xml.Name `xml:"live"`
		Streams []Stream `xml:"stream"`
	}

	// Stream is the individual stream element
	Stream struct {
		XMLName xml.Name `xml:"stream"`
		Name    string   `xml:"name"`
	}

	TransporterRouter struct {
		// ReturningChannel is the channel for returning data on
		ReturningChannel chan []byte
		// TransporterUnique is the payload to send to the client
		TransporterUnique transporter.TransporterUnique
	}

	RecallStream struct {
		StreamServer string `json:"streamServer"`
		StreamKey    string `json:"streamKey"`
	}

	ResumeResponse struct {
		Response  string `json:"response"`
		Error     string `json:"error"`
		Website   bool   `json:"website"`
		Recording bool   `json:"recording"`
		Streams   uint64 `json:"streams"`
	}

	StartSaveValidationResponse struct {
		Input           string
		RecordCheckbox  bool
		SavePath        string
		WebsiteCheckbox bool
		WebsiteOut      string
		Streams         []string
		Error           error
	}

	StatusResponse struct {
		Status []StatusResponseIndividual `json:"status"`
		Error  string                     `json:"error"`
	}

	StatusResponseIndividual struct {
		Name     string `json:"name"`
		Response string `json:"response"`
		Error    string `json:"error"`
	}

	ListedStream struct {
		Code  string `json:"code"`
		Input string `json:"input"`
	}

	// StartingType identifies which starting type this should act as
	StartingType int

	// ValidationType is used to determine how a form input should be validated
	ValidationType int
)

const (
	nonStoredStart StartingType = iota
	storedStart
)

const (
	startValidation ValidationType = iota
	startUniqueValidation
	saveValidation
)

const (
	charset                   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	internalChannelNameAppend = "Internal"
)

var (
	//nolint:gosec
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	wsUpgrade = websocket.Upgrader{}
)

// New initialises connections, templates, and cookies
func New(conf Config, store *store.Store) *Views {
	// Initialising session cookie
	authKey, _ := hex.DecodeString(conf.AuthenticationKey)
	if len(authKey) == 0 {
		authKey = securecookie.GenerateRandomKey(64)
	}

	encryptionKey, _ := hex.DecodeString(conf.EncryptionKey)
	if len(encryptionKey) == 0 {
		encryptionKey = securecookie.GenerateRandomKey(32)
	}

	cookie := sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)

	sixty := 60
	twentyFour := 24

	cookie.Options = &sessions.Options{
		MaxAge:   sixty * sixty * twentyFour,
		HttpOnly: true,
		Path:     "/",
	}

	return &Views{
		cache:    cache.New(cache.NoExpiration, 1*time.Hour),
		conf:     conf,
		cookie:   cookie,
		store:    store,
		template: templates.NewTemplate(conf.Version),
	}
}
