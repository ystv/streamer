package views

import (
	"log"
	"sync"
	"time"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
)

func (v *Views) BeginWatchdog() {
	log.Printf("watchdog initiailised")
	go func() {
		for {
			time.Sleep(15 * time.Second)
			if v.conf.Verbose {
				log.Printf("watchdog called")
			}
			streams, err := v.store.GetStreams()
			if err != nil {
				log.Printf("failed to get streams for watchdog: %+v", err)
				continue
			}

			_, rec := v.cache.Get(server.Recorder.String())
			_, fow := v.cache.Get(server.Forwarder.String())

			for _, stream := range streams {
				transporter := commonTransporter.Transporter{
					Action: action.Status,
					Unique: stream.Stream,
				}

				//nolint:staticcheck
				fStatus := commonTransporter.ForwarderStatus{
					Website: len(stream.Website) > 0,
					Streams: len(stream.Streams),
				}

				var forwarderError, recorderError bool

				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					if len(stream.Recording) > 0 && rec {
						recorderTransporter := transporter

						var response commonTransporter.ResponseTransporter
						response, err = v.wsHelper(server.Recorder, recorderTransporter)
						if err != nil {
							log.Printf("failed to send or receive message from recorder for status: %+v", err)
							recorderError = true
							return
						}
						if response.Status == wsMessages.Error {
							log.Printf("failed to get correct response from recorder for status: %s", response.Payload)
							recorderError = true
							return
						}
						if response.Status != wsMessages.Okay {
							log.Printf("invalid response from recorder for status: %s", response)
							recorderError = true
							return
						}
					}
				}()
				go func() {
					defer wg.Done()
					if fow {
						forwarderTransporter := transporter

						forwarderTransporter.Payload = fStatus

						var response commonTransporter.ResponseTransporter
						response, err = v.wsHelper(server.Forwarder, forwarderTransporter)
						if err != nil {
							log.Printf("failed to send or receive message from forwarder for status: %+v", err)
							forwarderError = true
							return
						}
						if response.Status == wsMessages.Error {
							log.Printf("failed to get correct response from forwarder for status: %s", response.Payload)
							forwarderError = true
							return
						}
						if response.Status != wsMessages.Okay {
							log.Printf("invalid response from recorder for status: %s", response)
							forwarderError = true
							return
						}
					}
				}()
				wg.Wait()

				if recorderError && rec {
					stopTransporter := commonTransporter.Transporter{
						Action: action.Stop,
						Unique: stream.Stream,
					}
					var wsResponse commonTransporter.ResponseTransporter
					wsResponse, err = v.wsHelper(server.Recorder, stopTransporter)
					if err != nil {
						log.Printf("failed sending to Recorder for stop: %+v", err)
					}
					if wsResponse.Status == wsMessages.Error {
						log.Printf("failed sending to Recorder for stop: %#v", wsResponse)
					}
					if wsResponse.Status != wsMessages.Okay {
						log.Printf("invalid response from Recorder for stop: %#v", wsResponse)
					}

					startTransporter := commonTransporter.Transporter{
						Action: action.Start,
						Unique: stream.Stream,
						Payload: commonTransporter.RecorderStart{
							StreamIn: stream.Input,
							PathOut:  stream.Recording,
						},
					}
					wsResponse, err = v.wsHelper(server.Recorder, startTransporter)
					if err != nil {
						log.Printf("failed sending to Recorder for start: %+v", err)
					}
					if wsResponse.Status == wsMessages.Error {
						log.Printf("failed sending to Recorder for start: %#v", wsResponse)
					}
					if wsResponse.Status != wsMessages.Okay {
						log.Printf("invalid response from Recorder for start: %#v", wsResponse)
					}
				}

				if forwarderError && fow {
					stopTransporter := commonTransporter.Transporter{
						Action: action.Stop,
						Unique: stream.Stream,
					}
					var wsResponse commonTransporter.ResponseTransporter
					wsResponse, err = v.wsHelper(server.Forwarder, stopTransporter)
					if err != nil {
						log.Printf("failed sending to Forwarder for stop: %+v", err)
					}
					if wsResponse.Status == wsMessages.Error {
						log.Printf("failed sending to Forwarder for stop: %#v", wsResponse)
					}
					if wsResponse.Status != wsMessages.Okay {
						log.Printf("invalid response from Forwarder for stop: %#v", wsResponse)
					}

					startTransporter := commonTransporter.Transporter{
						Action: action.Start,
						Unique: stream.Stream,
						Payload: commonTransporter.ForwarderStart{
							StreamIn:   stream.Input,
							WebsiteOut: stream.Website,
							Streams:    stream.Streams,
						},
					}
					wsResponse, err = v.wsHelper(server.Forwarder, startTransporter)
					if err != nil {
						log.Printf("failed sending to Forwarder for start: %+v", err)
					}
					if wsResponse.Status == wsMessages.Error {
						log.Printf("failed sending to Forwarder for start: %#v", wsResponse)
					}
					if wsResponse.Status != wsMessages.Okay {
						log.Printf("invalid response from Forwarder for start: %#v", wsResponse)
					}
				}
				time.Sleep(250 * time.Millisecond)
			}
		}
	}()
}
