package views

import (
	"fmt"
	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/helper/tx"
)

// StopFunc is used when the stream is finished,
// then you can stop the stream by pressing the stop button, and that would kill all the ffmpeg commands
func (v *Views) StopFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"list", http.StatusTemporaryRedirect)
		return
	}*/
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Stop POST called")
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			return fmt.Errorf("unique key invalid")
		}

		stream, err := v.store.FindStream(unique)
		if err != nil {
			return err
		}

		if stream == nil {
			return fmt.Errorf("no data in stream stop")
		}

		transporter := commonTransporter.Transporter{
			Action: action.Stop,
			Unique: unique,
		}

		var wg sync.WaitGroup
		_, rec := v.cache.Get(server.Recorder.String())
		_, fow := v.cache.Get(server.Forwarder.String())

		if (!rec && stream.Recording) && !fow {
			err = fmt.Errorf("no recorder or forwarder available")
		} else if !rec && stream.Recording {
			err = fmt.Errorf("no recorder available")
		} else if !fow {
			err = fmt.Errorf("no forwarder available")
		}
		if stream.Recording {
			wg.Add(2)
			go func() {
				defer wg.Done()
				//var client *ssh.Client
				//var session *ssh.Session
				//if recorderAuth == "PEM" {
				//	client, session, err = connectToHostPEM(recorder, recorderUsername, recorderPrivateKey, recorderPassphrase)
				//} else if recorderAuth == "PASS" {
				//client, session, err = helper.ConnectToHostPassword(v.conf.Recorder, v.conf.RecorderUsername, v.conf.RecorderPassword, v.conf.Verbose)
				//}
				//if err != nil {
				//	fmt.Println("Error connecting to Recorder for stop")
				//	fmt.Println(err)
				//}
				//_, err = session.CombinedOutput("./recorder_stop " + c.FormValue("unique") + " | bash")
				//if err != nil {
				//	fmt.Println("Error executing on Recorder for stop")
				//	fmt.Println(err)
				//}
				//err = client.Close()
				//if err != nil {
				//	fmt.Println(err)
				//}

				recorderTransporter := transporter

				var response commonTransporter.ResponseTransporter
				response, err = v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Println(err, "Error sending to Recorder for stop")
					return
				}
				if response.Status == wsMessages.Error {
					log.Printf("Error sending to Recorder for stop: %s", response)
					return
				}
				if response.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for stop: %s", response)
					return
				}

				fmt.Println("Recorder stop success")
			}()
		} else {
			wg.Add(1)
		}
		go func() {
			defer wg.Done()
			forwarderTransporter := transporter

			var response commonTransporter.ResponseTransporter
			response, err = v.wsHelper(server.Forwarder, forwarderTransporter)
			if err != nil {
				log.Println(err, "Error sending to Forwarder for stop")
				return
			}
			if response.Status == wsMessages.Error {
				log.Printf("Error sending to Forwarder for stop: %s", response)
				return
			}
			if response.Status != wsMessages.Okay {
				log.Printf("invalid response from Forwarder for stop: %s", response)
				return
			}

			fmt.Println("Forwarder stop success")
			//var client *ssh.Client
			//var session *ssh.Session
			//if forwarderAuth == "PEM" {
			//	client, session, err = connectToHostPEM(forwarder, forwarderUsername, forwarderPrivateKey, forwarderPassphrase)
			//} else if forwarderAuth == "PASS" {
			//client, session, err = helper.ConnectToHostPassword(v.conf.Forwarder, v.conf.ForwarderUsername, v.conf.ForwarderPassword, v.conf.Verbose)
			//}
			//if err != nil {
			//	fmt.Println("Error connecting to Forwarder for stop")
			//	fmt.Println(err)
			//}
			//_, err = session.CombinedOutput("./forwarder_stop " + c.FormValue("unique") + " | bash")
			//if err != nil {
			//	fmt.Println("Error executing on Forwarder for stop")
			//	fmt.Println(err)
			//}
			//err = client.Close()
			//if err != nil {
			//	fmt.Println(err)
			//}

			fmt.Println("Forwarder stop success")
		}()
		wg.Wait()
		fmt.Println("STOPPED!")

		err = v.store.DeleteStream(unique)
		if err != nil {
			return err
		}

		err = v.HandleTXLight(v.conf.TransmissionLight, tx.AllOff)
		if err != nil {
			fmt.Println(err)
		}

		return c.String(http.StatusOK, "STOPPED!")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
