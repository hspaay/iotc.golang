package iotconnect

// import (
// 	"flag"
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"runtime"
// 	"runtime/pprof"
// 	"syscall"

// 	log "github.com/sirupsen/logrus"
// )

// // RunPublisher Start the Publisher
// // * parse the commandline for config and debug level override
// // * load the configuration from the given folder
// // * Runs os.Exit(1) when completed with an error to let systemd restart it.
// func RunPublisher(starter func()) {
// 	//adapterConfigFile := path.Join(DefaultConfigFolder, configFile)
// 	configFolder := "./config"

// 	// parse commandline arguments. These override configuration file defaults
// 	configFolderPtr := flag.String("c", configFolder, "Adapter configuration folder")
// 	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
// 	memprofile := flag.String("memprofile", "", "write memory profile to `file`")
// 	flag.Parse()

// 	// Check config folder
// 	if _, err := os.Stat(*configFolderPtr); os.IsNotExist(err) {
// 		fmt.Printf("Configuration folder '%s' not found\n", *configFolderPtr)
// 		os.Exit(1)
// 	}

// 	log.Warningf("RunAdapter from config %s", *configFolderPtr)

// 	// CPU profiling output
// 	if *cpuprofile != "" {
// 		f, err := os.Create(*cpuprofile)
// 		if err != nil {
// 			log.Fatal("could not create CPU profile: ", err)
// 		}
// 		err = pprof.StartCPUProfile(f)
// 		defer pprof.StopCPUProfile()
// 	}

// 	err := adapter.LoadConfiguration(*configFolderPtr)
// 	if err != nil {
// 		os.Exit(2)
// 	}

// 	err = adapter.Start()
// 	if err != nil {
// 		log.Error("Failed starting adapter")
// 		os.Exit(1)
// 	}

// 	// catch all signals since not explicitly listing
// 	exitChannel := make(chan os.Signal, 1)
// 	//done := make(chan bool, 1)
// 	//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
// 	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)
// 	//signal.Notify(exitChannel, os.Interrupt)

// 	sig := <-exitChannel
// 	log.Warningf("RECEIVED SIGNAL: %s", sig)
// 	fmt.Println()
// 	fmt.Println(sig)
// 	adapter.Stop()

// 	// Memory profiling output
// 	if *memprofile != "" {
// 		log.Info("Creating memory profile in: ", memprofile)
// 		f, err := os.Create(*memprofile)
// 		if err != nil {
// 			log.Fatal("could not create memory profile: ", err)
// 		}
// 		runtime.GC() // get up-to-date statistics
// 		_ = pprof.WriteHeapProfile(f)
// 		_ = f.Close()
// 	}
// 	// wait for completion
// 	log.Info("Bye Bye...")
// 	// exit 1 so systemd restarts
// 	os.Exit(1)
// 	return
// }
