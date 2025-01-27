package views

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log" //nolint:gosec
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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
		Commit                string
		StreamServer          string `envconfig:"STREAM_SERVER"`
		TransmissionLight     string `envconfig:"TRANSMISSION_LIGHT"`
		APIEndpoint           string `envconfig:"API_ENDPOINT"`
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
		template: templates.NewTemplate(conf.Version, conf.Commit),
	}
}

func (v *Views) Authenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		data := struct {
			Error error `json:"error"`
		}{}
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Printf("failed to get session for authenticated: %+v", err)

			data.Error = fmt.Errorf("failed to get session for authenticated: %w", err)

			return c.JSON(http.StatusInternalServerError, data)
		}

		client := http.Client{Timeout: 2 * time.Second}

		var t struct {
			Token string `json:"token"`
		}
		var req *http.Request
		var resp *http.Response
		var b []byte

		token, ok := session.Values["token"].(string)
		if ok {
			req, err = http.NewRequestWithContext(c.Request().Context(), "GET",
				v.conf.AuthEndpoint+"/api/test", nil)
			if err != nil {
				log.Printf("failed to create new test token request: %+v", err)
				goto getToken
			}
			req.Header.Add("Authorization", "Bearer "+token)

			resp, err = client.Do(req)
			if err != nil {
				log.Printf("failed to do client for test token: %+v", err)
				goto getToken
			}
			defer resp.Body.Close()

			b, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("failed to read test token body: %+v", err)
				goto getToken
			}

			var response struct {
				StatusCode int    `json:"status_code"`
				Message    string `json:"message"`
			}
			err = json.Unmarshal(b, &response)
			if err != nil {
				log.Printf("failed to unmarshal JSON for test token: %+v", err)
				goto getToken
			}

			if response.StatusCode != 200 || resp.StatusCode != 200 || response.Message != "valid token" {
				goto getToken
			}

			return next(c)
		}

	getToken:
		req, err = http.NewRequestWithContext(c.Request().Context(), "GET",
			v.conf.AuthEndpoint+"/api/set_token", nil)
		if err != nil {
			log.Printf("failed to create new get token request: %+v", err)
			goto login
		}

		for _, cookie := range c.Request().Cookies() {
			req.AddCookie(cookie)
		}

		resp, err = client.Do(req)
		if err != nil {
			log.Printf("failed to do client for get token: %+v", err)
			goto login
		}
		defer resp.Body.Close()

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("failed to read get token body: %+v", err)
			goto login
		}

		err = json.Unmarshal(b, &t)
		if err != nil {
			log.Printf("failed to unmarshal JSON for get token: %+v", err)
			goto login
		}

		if t.Token == "" {
			goto login
		}

		if resp.StatusCode != 201 {
			goto login
		}

		if !hasCobra(t.Token) {
			return echo.NewHTTPError(http.StatusForbidden, errors.New("you are not authorised for accessing this site"))
		}

		session.Values["token"] = t.Token

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			log.Printf("failed to save token session for authentication: %+v", err)
			goto login
		}
		return next(c)

	login:
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?callback=https://%s%s",
			v.conf.AuthEndpoint,
			v.conf.StreamerWebAddress,
			c.Request().URL.String()))
	}
}

func hasCobra(token string) bool {
	split := strings.Split(token, ".")
	if len(split) != 3 {
		log.Printf("invalid token format: %s", token)
		return false
	}

	bytes, err := base64.RawURLEncoding.DecodeString(split[1])
	if err != nil {
		log.Printf("invalid token format: %s, %+v", token, err)
		return false
	}

	var perms struct {
		UserID      int      `json:"id"`
		Permissions []string `json:"perms"`
	}

	err = json.Unmarshal(bytes, &perms)
	if err != nil {
		log.Printf("invalid token format: %s, %+v", token, err)
		return false
	}

	for _, perm := range perms.Permissions {
		if perm == "COBRA" || perm == "SuperUser" {
			log.Printf("found Cobra or SuperUser permission for %d: %s", perms.UserID, perm)
			return true
		}
	}
	return false
}
