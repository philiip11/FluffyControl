package main

import (
	"github.com/chbmuc/lirc"
	"github.com/jasonlvhit/gocron"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
Available IR Codes for Samsung Navibot SR 8855
*/
const (
	POWER      = "POWER"
	RECHARGING = "RECHARGING"
	AUTO       = "AUTO"
	SPOT       = "SPOT"
	MAX        = "MAX"
	STARTSTOP  = "STARTSTOP"
	UP         = "UP"
	LEFT       = "LEFT"
	RIGHT      = "RIGHT"
	MANUAL     = "MANUAL"
	EDGE       = "EDGE"
	TIMERDAILY = "TIMERDAILY"
	CLOCK      = "CLOCK"
)

//change this if you prefer another mode
var defaultMode = EDGE

var running = false
var charging = false
var ir *lirc.Router
var lastRun = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

func sendIr(command string) {
	log.Println("Sende " + command + ".")
	err := ir.Send("fluffy " + command)
	if err != nil {
		log.Println(err)
	}
	if command == AUTO || command == SPOT || command == MAX || command == EDGE {
		lastRun = time.Now()
	}

}

func intelligentClean() {
	if lastRun.Add(8 * time.Hour).Before(time.Now()) {
		log.Println("Intelligentes Saugen: Akku wird geleert")
		running = true
		sendIr(defaultMode)
		time.Sleep(10 * time.Minute)
		log.Println("Intelligentes Saugen: Zeit zum Aufladen")
		sendIr(STARTSTOP)
		time.Sleep(20 * time.Minute)
		log.Println("Intelligentes Saugen: fertig geladen und los")
		sendIr(defaultMode)
	} else {
		log.Println("Intelligentes Saugen: nicht nötig, da Akku ausreichend geladen")
		sendIr(defaultMode)
	}
}

func start(w http.ResponseWriter, _ *http.Request) {
	log.Println("Starte den Staubsauger.")
	go intelligentClean()
	_, _ = io.WriteString(w, "OK\n")
}

func loop(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	hours, _ := strconv.Atoi(string(body))
	log.Println("Starte den Staubsauger im Dauerbetrieb für " + string(hours) + " Stunden.")
	go vacuum(hours)
	_, _ = io.WriteString(w, "OK\n")
}

func vacuum(hours int) {
	running = true
	for i := 0; i < hours; i++ {
		go intelligentClean()
		time.Sleep(1 * time.Hour)
		if !running {
			break
		}
	}
}

func setCron(t string) {
	_ = gocron.Every(1).Day().At(t).Do(intelligentClean)
}

func setTime(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	t := string(body)
	t = strings.ReplaceAll(t, "Uhr", "")
	log.Println("Starte den Staubsauger jeden Tag um " + t + " Uhr.")
	setCron(t)
	_, _ = io.WriteString(w, "OK\n")
}

func stop(w http.ResponseWriter, _ *http.Request) {
	log.Println("Stoppe den Staubsauger.")
	running = false
	sendIr(STARTSTOP)
	go func() {
		time.Sleep(5 * time.Second)
		sendIr(RECHARGING)
	}()
	_, _ = io.WriteString(w, "OK\n")
}

func showLog(w http.ResponseWriter, r *http.Request) {
	log.Println("Showing logfile to " + GetIP(r))
	logfile, _ := ioutil.ReadFile("/home/pi/go/fluffy-log.txt")
	_, _ = io.WriteString(w, string(logfile))

}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func main() {
	// INIT LOGGER
	file, err := os.OpenFile("/home/pi/go/fluffy-log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	defer file.Close()
	log.Println("Start Fluffy service")
	defer log.Println("Stopping Fluffy service")

	// INIT IR
	ir, err = lirc.Init("/var/run/lirc/lircd")
	if err != nil {
		log.Fatal(err)
	}

	// Restart Fluffy
	sendIr(POWER)
	time.Sleep(3 * time.Second)
	sendIr(POWER)

	http.HandleFunc("/Fluffy/start", start)
	http.HandleFunc("/Fluffy/stop", stop)
	http.HandleFunc("/Fluffy/loop", loop)
	http.HandleFunc("/Fluffy/log", showLog)
	http.HandleFunc("/Fluffy/setTime", setTime)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
