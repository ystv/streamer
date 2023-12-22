package views

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/helper"
	"golang.org/x/crypto/ssh"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// StatusFunc is used to check the status of the streams and does this by tail command of the output logs
func (v *Views) StatusFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Status POST called")
		}

		stream, err := v.store.FindStream(c.FormValue("unique"))
		if err != nil {
			return err
		}

		m := make(map[string]string)
		var wg sync.WaitGroup
		if stream != nil {
			if stream.Recording {
				wg.Add(2)
				go func() {
					defer wg.Done()
					var client *ssh.Client
					var session *ssh.Session
					var err error
					//if recorderAuth == "PEM" {
					//	client, session, err = connectToHostPEM(recorder, recorderUsername, recorderPrivateKey, recorderPassphrase)
					//} else if recorderAuth == "PASS" {
					client, session, err = helper.ConnectToHostPassword(v.conf.Recorder, v.conf.RecorderUsername, v.conf.RecorderPassword, v.conf.Verbose)
					//}
					if err != nil {
						fmt.Println("Error connecting to Recorder for status")
						fmt.Println(err)
					}
					dataOut, err := session.CombinedOutput("./recorder_status.sh " + c.FormValue("unique"))
					if err != nil {
						fmt.Println("Error executing on Recorder for status")
						fmt.Println(err)
					}
					err = client.Close()
					if err != nil {
						fmt.Println(err)
					}

					dataOut1 := string(dataOut)[:len(dataOut)-2]

					if len(dataOut1) > 0 {
						if strings.Contains(dataOut1, "frame=") {
							first := strings.Index(dataOut1, "frame=") - 1
							last := strings.LastIndex(dataOut1, "\r")
							dataOut1 = dataOut1[:last]
							last = strings.LastIndex(dataOut1, "\r") + 1
							m["recording"] = dataOut1[:first] + "\n" + dataOut1[last:]
						} else {
							m["recording"] = dataOut1
						}
					}

					fmt.Println("Recorder status success")
				}()
			} else {
				wg.Add(1)
			}
			go func() {
				defer wg.Done()
				var client *ssh.Client
				var session *ssh.Session
				var err error
				//if forwarderAuth == "PEM" {
				//	client, session, err = connectToHostPEM(forwarder, forwarderUsername, forwarderPrivateKey, forwarderPassphrase)
				//} else if forwarderAuth == "PASS" {
				client, session, err = helper.ConnectToHostPassword(v.conf.Forwarder, v.conf.ForwarderUsername, v.conf.ForwarderPassword, v.conf.Verbose)
				//}
				if err != nil {
					fmt.Println("Error connecting to Forwarder for status")
					fmt.Println(err)
				}
				dataOut, err := session.CombinedOutput("./forwarder_status " + strconv.FormatBool(stream.Website) + " " + strconv.FormatUint(stream.Streams, 10) + " " + c.FormValue("unique"))
				if err != nil {
					fmt.Println("Error executing on Forwarder for status")
					fmt.Println(err)
				}
				err = client.Close()
				if err != nil {
					fmt.Println(err)
				}

				dataOut1 := string(dataOut)[4 : len(dataOut)-2]

				dataOut2 := strings.Split(dataOut1, "\u0000")

				for _, dataOut3 := range dataOut2 {
					if len(dataOut3) > 0 {
						if strings.Contains(dataOut3, "frame=") {
							dataOut4 := strings.Split(dataOut3, "~:~")
							first := strings.Index(dataOut4[1], "frame=") - 1
							last := strings.LastIndex(dataOut4[1], "\r")
							dataOut4[1] = dataOut4[1][:last]
							last = strings.LastIndex(dataOut4[1], "\r") + 1
							if len(dataOut4) > 1 {
								m[strings.Trim(dataOut4[0], " ")] = dataOut4[1][:first] + "\n" + dataOut4[1][last:]
							} else {
								fmt.Println(dataOut4)
								m[strings.Trim(dataOut4[0], " ")] = ""
							}
						} else {
							dataOut4 := strings.Split(dataOut3, "~:~")
							if len(dataOut4) > 1 {
								m[strings.Trim(dataOut4[0], " ")] = dataOut4[1]
							} else {
								fmt.Println(dataOut4)
								m[strings.Trim(dataOut4[0], " ")] = ""
							}
						}
					}
				}

				fmt.Println("Forwarder status success")
			}()
			wg.Wait()
			var jsonStr []byte
			jsonStr, err = json.Marshal(m)
			if err != nil {
				return err
			}
			output := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(jsonStr[1:len(jsonStr)-1]), "\\n", "<br>"), "\"", ""), " , ", "<br><br><br>"), " ,", "<br><br><br>"), "<br>,", "<br><br>")
			return c.String(http.StatusOK, output)
		}
		return fmt.Errorf("ERROR DATA STATUS")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
