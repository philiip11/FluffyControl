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
	//POWER      = "POWER"
	RECHARGING = "RECHARGING"
	AUTO       = "AUTO"
	SPOT       = "SPOT"
	MAX        = "MAX"
	STARTSTOP  = "STARTSTOP"
	//UP         = "UP"
	//LEFT       = "LEFT"
	//RIGHT      = "RIGHT"
	//MANUAL     = "MANUAL"
	EDGE = "EDGE"
	//TIMERDAILY = "TIMERDAILY"
	//CLOCK = "CLOCK"
)

//change this if you prefer another mode
var defaultMode = EDGE

var running = false
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
		if !running {
			log.Println("Intelligentes Saugen: abgebrochen")
			return
		}
		log.Println("Intelligentes Saugen: Zeit zum Aufladen")
		sendIr(STARTSTOP)
		time.Sleep(20 * time.Minute)
		if !running {
			log.Println("Intelligentes Saugen: abgebrochen")
			return
		}
		log.Println("Intelligentes Saugen: fertig geladen und los")
		sendIr(defaultMode)
	} else {
		log.Println("Intelligentes Saugen: nicht nötig, da Akku ausreichend geladen")
		sendIr(defaultMode)
	}
}

func start(w http.ResponseWriter, r *http.Request) {
	log.Println("Starte den Staubsauger. IP: " + GetIP(r))
	go intelligentClean()
	_, _ = io.WriteString(w, "OK\n")
}

func stop(w http.ResponseWriter, r *http.Request) {
	log.Println("Stoppe den Staubsauger. IP: " + GetIP(r))
	running = false
	sendIr(STARTSTOP)
	go func() {
		time.Sleep(5 * time.Second)
		sendIr(RECHARGING)
	}()
	_, _ = io.WriteString(w, "OK\n")
}
func loop(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	hours, _ := strconv.Atoi(string(body))
	log.Println("Starte den Staubsauger im Dauerbetrieb für " + string(hours) + " Stunden. IP: " + GetIP(r))
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
	log.Println("Starte den Staubsauger jeden Tag um " + t + " Uhr.")
	gocron.Clear()
	err := gocron.Every(1).Day().At(t).Do(intelligentClean)
	if err != nil {
		log.Println("gocron: " + err.Error())
	}
	go func() { <-gocron.Start() }()
	_ = ioutil.WriteFile("/home/pi/go/gocron", []byte(t), 0644)
}

func setTime(w http.ResponseWriter, r *http.Request) {
	log.Println("IP: " + GetIP(r))
	body, _ := ioutil.ReadAll(r.Body)
	t := string(body)
	t = strings.ReplaceAll(t, "Uhr", "")
	t = strings.ReplaceAll(t, "uhr", "")
	t = strings.ReplaceAll(t, "ur", "")
	t = strings.ReplaceAll(t, " ", "")
	if !strings.Contains(t, ":") && len(t) == 2 {
		t = t + ":00"
	}
	setCron(t)
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
	log.Println("####################")
	log.Println("Start Fluffy service")
	defer log.Println("Stopping Fluffy service")

	// INIT IR
	ir, err = lirc.Init("/var/run/lirc/lircd")
	if err != nil {
		log.Fatal(err)
	}

	// INIT gocron
	if _, err := os.Stat("/home/pi/go/gocron"); err == nil {
		readFile, err := ioutil.ReadFile("/home/pi/go/gocron")
		if err != nil {
			log.Fatal(err)
		}
		setCron(string(readFile))
	}

	// This is not important but it makes a sound so I know when this program starts
	sendIr(RECHARGING)

	http.HandleFunc("/Fluffy/start", start)
	http.HandleFunc("/Fluffy/stop", stop)
	http.HandleFunc("/Fluffy/loop", loop)
	http.HandleFunc("/Fluffy/log", showLog)
	http.HandleFunc("/Fluffy/setTime", setTime)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
