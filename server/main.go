package main

import (
	//"encoding/pem"
	"encoding/xml"
	//"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ystv/streamer/server/templates"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/mattn/go-sqlite3"
)

type (
	Web struct {
		mux *mux.Router
		t   *templates.Templater
		cfg Config
	}

	Config struct {
		Forwarder         string `envconfig:"FORWARDER"`
		Recorder          string `envconfig:"RECORDER"`
		ForwarderUsername string `envconfig:"FORWARDER_USERNAME"`
		RecorderUsername  string `envconfig:"RECORDER_USERNAME"`
		ForwarderPassword string `envconfig:"FORWARDER_PASSWORD"`
		RecorderPassword  string `envconfig:"RECORDER_PASSWORD"`
		StreamChecker     string `envconfig:"STREAM_CHECKER"`
		TransmissionLight string `envconfig:"TRANSMISSION_LIGHT"`
		KeyChecker        string `envconfig:"KEY_CHECKER"`
		ServerPort        int    `envconfig:"SERVER_PORT"`
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

	/*Claims struct {
		Id    int          `json:"id"`
		Perms []Permission `json:"perms"`
		Exp   int64        `json:"exp"`
		jwt.StandardClaims
	}*/

	/*Permission struct {
		Permission string `json:"perms"`
		jwt.StandardClaims
	}*/

	/*Views struct {
		cookie *sessions.CookieStore
	}*/
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var verbose bool

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// main function is the start and the root for the website
func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./streamer") {
		if len(os.Args) > 2 {
			fmt.Println(string(rune(len(os.Args))))
			fmt.Println(os.Args)
			log.Fatalf("Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) > 1 {
			fmt.Println(string(rune(len(os.Args))))
			fmt.Println(os.Args)
			log.Fatalf("Arguments error")
		}
	}
	if os.Args[0] == "-v" {
		verbose = true
	} else {
		verbose = false
	}

	err := godotenv.Load()
	if err != nil {
		log.Printf("error loading .env file: %s", err)
	}

	var cfg Config
	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("failed to process env vars: %s", err)
	}

	web := Web{
		mux: mux.NewRouter(),
		cfg: cfg,
	}
	web.mux.HandleFunc("/", web.home)
	//web.mux.HandleFunc("/authenticate1", web.authenticate1)
	web.mux.HandleFunc("/endpoints", web.endpoints)
	web.mux.HandleFunc("/streams", web.streams)
	web.mux.HandleFunc("/start", web.start)
	web.mux.HandleFunc("/resume", web.resume)
	web.mux.HandleFunc("/status", web.status)
	web.mux.HandleFunc("/stop", web.stop)
	web.mux.HandleFunc("/list", web.list)
	web.mux.HandleFunc("/save", web.save)
	web.mux.HandleFunc("/recall", web.recall)
	web.mux.HandleFunc("/youtubehelp", web.youtubeHelp)
	web.mux.HandleFunc("/facebookhelp", web.facebookHelp)
	web.mux.HandleFunc("/public/{id:[a-zA-Z0-9_.-]+}", web.public) // This handles all the public pages that the webpage can request, e.g. css, images and jquery

	fmt.Println("Server listening on port", web.cfg.ServerPort, "...")

	err = http.ListenAndServe(net.JoinHostPort("", strconv.Itoa(web.cfg.ServerPort)), web.mux)

	if err != nil {
		fmt.Println(err)
		return
	}
}

/*func authenticate(w http.ResponseWriter, r *http.Request) bool {
	_ = w
	response, err := http.Get("https://auth.dev.ystv.co.uk/api/set_token")
	if err != nil {
		fmt.Println(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(response.Body)
	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf.String())
	reqToken := r.Header.Get("Authorization")
	//splitToken := strings.Split(reqToken, "Bearer ")
	//reqToken = splitToken[1]
	fmt.Println("Token - ", reqToken)
	err = godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}

	sess := session.Get(r)
	if sess == nil {
		fmt.Println("None")
	} else {
		fmt.Println(sess)
	}

	jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

	http.Redirect(w, r, jwtAuthentication+"authenticate1", http.StatusTemporaryRedirect)
	return false
}

//
func (web *Web) authenticate1(w http.ResponseWriter, r *http.Request) {
	_ = w
	reqToken := r.Header.Get("Authorization")
	//splitToken := strings.Split(reqToken, "Bearer ")
	//reqToken = splitToken[1]
	fmt.Println("Token - ", reqToken)
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}
	jwtKey := os.Getenv("JWT_KEY")

	//fmt.Println(r.Cookies())

	fmt.Println(r)

	view := Views{}

	view.cookie = sessions.NewCookieStore(
		[]byte("444bd23239f14b804af0ae40375c8feec80b699684f4d1a6d86f59658edb3706caaa306fd3361e6353bf54c0df66adb7c1e395cac79a72ee0339dc1892fd478e"),
		[]byte("444bd23239f14b804af0ae40375c8feec80b699684f4d1a6d86f59658edb3706"),
	)

	sess, err := view.cookie.Get(r, "session")

	fmt.Println(sess)

	_ = sess

	//store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	//session, err := store.Get(r, "session-name")

	//fmt.Println(err)

	//fmt.Println(session.ID)

	//fmt.Println(session)

	//fmt.Println(session.Values["token"])

	response, err := http.Get("https://auth.dev.ystv.co.uk/api/set_token")
	if err != nil {
		fmt.Println(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(response.Body)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		fmt.Println(err)
	}

	tokenPage := buf.String()

	_ = tokenPage

	//fmt.Println(tokenPage)

	fmt.Println(r.Cookie("session"))

	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			fmt.Println(err)

		}
		fmt.Println(err)
		return
	}

	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	fmt.Println(tkn)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			fmt.Println("Unauthorised")
			fmt.Println(err)
			return
		}
		fmt.Println(err)
		return
	}
	if !tkn.Valid {
		fmt.Println("Unauthorised")
		return
	}
	if time.Now().Unix() > claims.Exp {
		fmt.Println("Expired")
		return
	}
	for _, perm := range claims.Perms {
		if perm.Permission == "Streamer" {
			fmt.Println("~~~Success~~~")
			return
		}
	}
	fmt.Println("Unauthorised")
	return
}









func connectToHostPEM(host, username, privateKeyPath, privateKeyPassword string) (*ssh.Client, *ssh.Session, error) {
	pemBytes, err := ioutil.ReadFile(privateKeyPath)
	signer, err := signerFromPem(pemBytes, []byte(privateKeyPassword))
	if err != nil {
		return nil, nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		err := client.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	return client, session, nil
}

func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {

	// read pem block
	err := errors.New("Pem decode failed, no key found")
	pemBlock, _ := pem.Decode(pemBytes)
	if pemBlock == nil {
		return nil, err
	}

	// handle encrypted key
	if x509.IsEncryptedPEMBlock(pemBlock) {
		// decrypt PEM
		pemBlock.Bytes, err = x509.DecryptPEMBlock(pemBlock, password)
		if err != nil {
			return nil, fmt.Errorf("Decrypting PEM block failed %v", err)
		}

		// get RSA, EC or DSA key
		key, err := parsePemBlock(pemBlock)
		if err != nil {
			return nil, err
		}

		// generate signer instance from key
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, fmt.Errorf("Creating signer from encrypted key failed %v", err)
		}

		return signer, nil
	} else {
		// generate signer instance from plain key
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing plain private key failed %v", err)
		}

		return signer, nil
	}
}

func parsePemBlock(block *pem.Block) (interface{}, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing PKCS private key failed %v", err)
		} else {
			return key, nil
		}
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing EC private key failed %v", err)
		} else {
			return key, nil
		}
	case "DSA PRIVATE KEY":
		key, err := ssh.ParseDSAPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("Parsing DSA private key failed %v", err)
		} else {
			return key, nil
		}
	default:
		return nil, fmt.Errorf("Parsing private key failed, unsupported key type %q", block.Type)
	}
}*/
