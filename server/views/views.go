package views

import (
	"encoding/xml"
	"github.com/ystv/streamer/server/store"
	"github.com/ystv/streamer/server/templates"
	"math/rand"
	"time"
)

type (
	// Config the global web-auth configuration
	Config struct {
		Verbose           bool
		Forwarder         string `envconfig:"FORWARDER"`
		Recorder          string `envconfig:"RECORDER"`
		ForwarderUsername string `envconfig:"FORWARDER_USERNAME"`
		RecorderUsername  string `envconfig:"RECORDER_USERNAME"`
		ForwarderPassword string `envconfig:"FORWARDER_PASSWORD"`
		RecorderPassword  string `envconfig:"RECORDER_PASSWORD"`
		StreamServer      string `envconfig:"STREAM_SERVER"`
		TransmissionLight string `envconfig:"TRANSMISSION_LIGHT"`
		KeyChecker        string `envconfig:"KEY_CHECKER"`
		ServerPort        int    `envconfig:"SERVER_PORT"`
		ServerAddress     string `envconfig:"SERVER_ADDRESS"`
		RecordingLocation string `envconfig:"RECORDING_LOCATION"`
	}

	// Views encapsulates our view dependencies
	Views struct {
		conf     Config
		template *templates.Templater
		store    *store.Store
	}

	RTMP struct {
		XMLName xml.Name `xml:"rtmp"`
		Server  Server   `xml:"server"`
	}

	Server struct {
		XMLName      xml.Name      `xml:"server"`
		Applications []Application `xml:"application"`
	}

	Application struct {
		XMLName xml.Name `xml:"application"`
		Name    string   `xml:"name"`
		Live    Live     `xml:"live"`
	}

	Live struct {
		XMLName xml.Name `xml:"live"`
		Streams []Stream `xml:"stream"`
	}

	Stream struct {
		XMLName xml.Name `xml:"stream"`
		Name    string   `xml:"name"`
	}
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// New initialises connections, templates, and cookies
func New(conf Config, store *store.Store) *Views {
	v := &Views{}

	v.template = templates.NewTemplate()

	v.conf = conf

	v.store = store

	return v
}
