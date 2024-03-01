package views

import (
	"encoding/xml"
	"math/rand"
	"time"

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
		Forwarder             string `envconfig:"FORWARDER"`
		Recorder              string `envconfig:"RECORDER"`
		ForwarderUsername     string `envconfig:"FORWARDER_USERNAME"`
		RecorderUsername      string `envconfig:"RECORDER_USERNAME"`
		ForwarderPassword     string `envconfig:"FORWARDER_PASSWORD"`
		RecorderPassword      string `envconfig:"RECORDER_PASSWORD"`
		StreamServer          string `envconfig:"STREAM_SERVER"`
		TransmissionLight     string `envconfig:"TRANSMISSION_LIGHT"`
		KeyChecker            string `envconfig:"KEY_CHECKER"`
		ServerPort            int    `envconfig:"SERVER_PORT"`
		ServerAddress         string `envconfig:"SERVER_ADDRESS"`
		RecordingLocation     string `envconfig:"RECORDING_LOCATION"`
		StreamerWebsocketPath string `envconfig:"STREAMER_WEBSOCKET_PATH"`
		StreamerAdminPath     string `envconfig:"STREAMER_ADMIN_PATH"`
	}

	// Views encapsulates our view dependencies
	Views struct {
		cache    *cache.Cache
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

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// New initialises connections, templates, and cookies
func New(conf Config, store *store.Store) *Views {
	return &Views{
		cache:    cache.New(cache.NoExpiration, 1*time.Hour),
		conf:     conf,
		store:    store,
		template: templates.NewTemplate(conf.Version),
	}
}
