package views

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
)

// StatusFunc is used to check the status of the streams and does this by tail command of the output logs
func (v *Views) StatusFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Status POST called")
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			return fmt.Errorf("unique key invalid")
		}

		stream, err := v.store.FindStream(unique)
		if err != nil {
			return fmt.Errorf("unable to find stream for status: %s, %w", unique, err)
		}

		if stream == nil {
			return fmt.Errorf("failed to get stream as data is empty")
		}

		transporter := commonTransporter.Transporter{
			Action: action.Status,
			Unique: unique,
		}

		fStatus := commonTransporter.ForwarderStatus{
			Website: stream.Website,
			Streams: int(stream.Streams),
		}

		m := make(map[string]string)
		var wg sync.WaitGroup
		if stream != nil {
			if stream.Recording {
				wg.Add(2)
				go func() {
					defer wg.Done()
					recorderTransporter := transporter

					var response commonTransporter.ResponseTransporter
					response, err = v.wsHelper(server.Recorder, recorderTransporter)
					if err != nil {
						log.Println(err, "Error sending to Recorder for status")
						return
					}
					if response.Status == wsMessages.Error {
						log.Printf("Error sending to Recorder for status: %s", response)
						return
					}
					if response.Status != wsMessages.Okay {
						log.Printf("invalid response from Recorder for status: %s", response)
						return
					}
					m["recording"] = response.Payload.(string)

					fmt.Println("Recorder status success")
				}()
			} else {
				wg.Add(1)
			}
			go func() {
				defer wg.Done()
				forwarderTransporter := transporter

				forwarderTransporter.Payload = fStatus

				var response commonTransporter.ResponseTransporter
				response, err = v.wsHelper(server.Forwarder, forwarderTransporter)
				if err != nil {
					log.Println(err, "Error sending to Recorder for status")
					return
				}
				if response.Status == wsMessages.Error {
					log.Printf("Error sending to Recorder for status: %s", response)
					return
				}
				if response.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for status: %s", response)
					return
				}

				var forwarderStatus commonTransporter.ForwarderStatusResponse

				err = mapstructure.Decode(response.Payload, &forwarderStatus)
				if err != nil {
					log.Printf("failed to decode: %+v", err)
					return
				}

				if len(forwarderStatus.Website) > 0 {
					m["website"] = forwarderStatus.Website
				}

				for index, streamOut := range forwarderStatus.Streams {
					m[strconv.Itoa(int(index))] = streamOut
				}

				//var client *ssh.Client
				//var session *ssh.Session
				//var err error
				//if forwarderAuth == "PEM" {
				//	client, session, err = connectToHostPEM(forwarder, forwarderUsername, forwarderPrivateKey, forwarderPassphrase)
				//} else if forwarderAuth == "PASS" {
				//client, session, err = helper.ConnectToHostPassword(v.conf.Forwarder, v.conf.ForwarderUsername, v.conf.ForwarderPassword, v.conf.Verbose)
				//}
				//if err != nil {
				//	fmt.Println("Error connecting to Forwarder for status")
				//	fmt.Println(err)
				//}
				//var dataOut []byte
				//dataOut, err = session.CombinedOutput("./forwarder_status " + strconv.FormatBool(stream.Website) + " " + strconv.FormatUint(stream.Streams, 10) + " " + c.FormValue("unique"))
				//if err != nil {
				//	fmt.Println("Error executing on Forwarder for status")
				//	fmt.Println(err)
				//}
				//err = client.Close()
				//if err != nil {
				//	fmt.Println(err)
				//}
				//
				//dataOut1 := string(dataOut)[4 : len(dataOut)-2]
				//
				//dataOut2 := strings.Split(dataOut1, "\u0000")
				//
				//for _, dataOut3 := range dataOut2 {
				//	if len(dataOut3) > 0 {
				//		if strings.Contains(dataOut3, "frame=") {
				//			dataOut4 := strings.Split(dataOut3, "~:~")
				//			first := strings.Index(dataOut4[1], "frame=") - 1
				//			last := strings.LastIndex(dataOut4[1], "\r")
				//			dataOut4[1] = dataOut4[1][:last]
				//			last = strings.LastIndex(dataOut4[1], "\r") + 1
				//			if len(dataOut4) > 1 {
				//				m[strings.Trim(dataOut4[0], " ")] = dataOut4[1][:first] + "\n" + dataOut4[1][last:]
				//			} else {
				//				fmt.Println(dataOut4)
				//				m[strings.Trim(dataOut4[0], " ")] = ""
				//			}
				//		} else {
				//			dataOut4 := strings.Split(dataOut3, "~:~")
				//			if len(dataOut4) > 1 {
				//				m[strings.Trim(dataOut4[0], " ")] = dataOut4[1]
				//			} else {
				//				fmt.Println(dataOut4)
				//				m[strings.Trim(dataOut4[0], " ")] = ""
				//			}
				//		}
				//	}
				//}

				fmt.Println("Forwarder status success")
			}()
			wg.Wait()
			var jsonStr []byte
			jsonStr, err = json.Marshal(m)
			if err != nil {
				return err
			}
			output := strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						strings.ReplaceAll(
							strings.ReplaceAll(
								string(jsonStr[1:len(jsonStr)-1]), "\\n", "<br>"),
							"\"", ""),
						" , ", "<br><br><br>"),
					" ,", "<br><br><br>"),
				"<br>,", "<br><br>")
			return c.String(http.StatusOK, output)
		}
		return fmt.Errorf("ERROR DATA STATUS")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
