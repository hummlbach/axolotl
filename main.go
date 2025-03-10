package main

import (
	"flag"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	astilectron "github.com/asticode/go-astilectron"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/nanu-c/axolotl/app/config"
	"github.com/nanu-c/axolotl/app/helpers"
	"github.com/nanu-c/axolotl/app/push"
	"github.com/nanu-c/axolotl/app/ui"
	"github.com/nanu-c/axolotl/app/webserver"
	"github.com/nanu-c/axolotl/app/worker"
)

func init() {
	flag.StringVar(&config.MainQml, "qml", "qml/phoneui/main.qml", "The qml file to load.")
	flag.StringVar(&config.Gui, "e", "", "Specify runtime environment. Use either electron, ut, lorca, qt or server")
	flag.StringVar(&config.AxolotlWebDir, "axolotlWebDir", "./axolotl-web/dist", "Specify the directory to use for axolotl-web")
	flag.BoolVar(&config.ElectronDebug, "eDebug", false, "Open electron development console")
	flag.BoolVar(&config.PrintVersion, "version", false, "Print version info")
	flag.StringVar(&config.ServerHost, "host", "127.0.0.1", "Host to serve UI from.")
	flag.StringVar(&config.ServerPort, "port", "9080", "Port to serve UI from.")
}
func setup() {
	helpers.SetupLogging()
	config.SetupConfig()
	log.SetLevel(log.DebugLevel)
	log.Infoln("[axolotl] Starting Signal for Ubuntu version", config.AppVersion)
}
func runBackend() {
	go worker.RunBackend()
	if config.IsPushHelper {
		push.PushHelperProcess()
	}
}
func runUI() error {
	defer wg.Done()
	if config.Gui != "ut" && config.Gui != "lorca" && config.Gui != "qt" {
		ui.RunUi(config.Gui)
		runElectron()
	} else {
		ui.RunUi(config.Gui)
	}
	os.Exit(0)
	return nil
}
func runElectron() {
	log.Infoln("[axolotl] Start electron")
	l := log.New()
	electronPath := os.Getenv("SNAP_USER_DATA")
	if len(electronPath) == 0 {
		electronPath = config.ConfigDir + "/electron"
	}
	var a, err = astilectron.New(l, astilectron.Options{
		AppName:            "axolotl",
		AppIconDefaultPath: "axolotl-web/public/axolotl.png", // If path is relative, it must be relative to the data directory
		AppIconDarwinPath:  "axolotl-web/public/axolotl.png", // Same here
		BaseDirectoryPath:  electronPath,
		VersionElectron:    "12.0.0",
		VersionAstilectron: "0.45.0",
		SingleInstance:     true,
		ElectronSwitches:   []string{"--disable-dev-shm-usage", "--no-sandbox"}})

	if err != nil {
		log.Errorln(errors.Wrap(err, "[axolotl-electron]: creating astilectron failed"))
	}

	defer a.Close()

	// Start astilectron
	a.HandleSignals()

	if err = a.Start(); err != nil {
		log.Errorln(errors.Wrap(err, "[axolotl-electron] main: starting astilectron failed"))
	}

	a.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App has crashed", e)
		return
	})
	a.HandleSignals()
	// New window
	var w *astilectron.Window
	center := true
	height := 800
	width := 600
	if w, err = a.NewWindow("http://"+config.ServerHost+":"+config.ServerPort, &astilectron.WindowOptions{
		Center: &center,
		Height: &height,
		Width:  &width,
	}); err != nil {
		log.Debugln("[axolotl-electron]", errors.Wrap(err, "main: new window failed"))
	}
	w.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
		log.Errorln("[axolotl-electron] Electron App has crashed")
		return
	})
	w.On(astilectron.EventNameWindowEventDidFinishLoad, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Electron App load")
		w.ExecuteJavaScript("window.onToken = function(token){window.location='http://" + config.ServerHost + ":" + config.ServerPort + "/?token='+token;};")
		return
	})
	w.On(astilectron.EventNameWindowEventWillNavigate, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Electron navigation")
		return
	})
	w.On(astilectron.EventNameWindowEventWebContentsExecutedJavaScript, func(e astilectron.Event) (deleteListener bool) {
		log.Infoln("[axolotl-electron] Electron navigation js")
		return
	})
	// Create windows
	if err = w.Create(); err != nil {
		log.Debugln("[axolotl-electron]", errors.Wrap(err, "main: creating window failed"))
	}
	log.Debugln("[axolotl-electron] open dev tools", config.ElectronDebug)

	if config.ElectronDebug {
		w.OpenDevTools()
	}
	w.Session.ClearCache()
	// Blocking pattern
	a.Wait()
}
func runWebserver() {
	defer wg.Done()
	log.Printf("[axolotl] Axolotl server started")
	// Fetch the URL.
	webserver.Run()
}

var wg sync.WaitGroup

func main() {
	setup()
	runBackend()
	log.Println("[axolotl] Setup completed")
	wg.Add(1)
	go runWebserver()
	if config.Gui != "server" {
		wg.Add(1)
		go runUI()
	} else {
		log.Printf("[axolotl] Axolotl frontend is at http://" + config.ServerHost + ":" + config.ServerPort + "/")
	}
	wg.Wait()
}
