//
//   date  : 2014-06-04
//   author: xjdrew
//

package main

import (
    "fmt"
    "flag"
    "io"

    "log"
    // "log/syslog"
    "os"
    "os/signal"
    "sync"
    "syscall"
)

var logger *log.Logger
func init() {
    // var err error
    //logger, err = syslog.NewLogger(syslog.LOG_LOCAL0, 0)
    logger = log.New(io.Writer(os.Stderr), "", 0)
    
    /*
    if err != nil {
        fmt.Printf("create logger failed:%s", err.Error())
        os.Exit(1)
    }
    */
    logger.Println("gotunnel run!")
}

type Service interface {
    Start() error
    Stop()
    Wait()
}

type App struct {
    services []Service
    wg sync.WaitGroup
}

func (self *App) Add(service Service) {
    self.services = append(self.services, service)
}

func (self *App) Start() error {
    for _, service := range self.services {
        err := service.Start()
        if err != nil {
            return err
        }
    }

    for _, service := range self.services {
        self.wg.Add(1)
        go func(s Service) {
            defer self.wg.Done()
            s.Wait()
            logger.Printf("service finish: %v", s)
        }(service)
    }
    return nil
}

func (self *App) Stop() {
    for _, service := range self.services {
        service.Stop()
    }
}

func (self *App) Wait() {
    self.wg.Wait()
}

const SIG_STOP   = syscall.Signal(34)
const SIG_STATUS = syscall.Signal(35)

func handleSignal(app *App) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, SIG_STOP, SIG_STATUS, syscall.SIGTERM)

    for sig := range(c) {
        switch sig {
            case SIG_STOP:
                app.Stop()
            case SIG_STATUS:
                logger.Println("catch sigstatus, ignore")
            case syscall.SIGTERM:
                logger.Println("catch sigterm, ignore")
        }
    }
}

type Options struct {
    gate bool
    capacity int
    frontAddr string
    backAddr string
    configFile string
}
var options Options

func usage() {
    fmt.Fprintf(os.Stderr, "usage: %s [configFile]\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(1)
}

func main() {
    flag.BoolVar(&options.gate, "gate", false, "as gate or node")
    flag.IntVar(&options.capacity, "capacity", 0xffff, "max concurrent connections(65535)")
    flag.StringVar(&options.frontAddr, "front_addr", "0.0.0.0:8001", "front door address(0.0.0.0:8001)")
    flag.StringVar(&options.backAddr, "back_addr", "0.0.0.0:8002", "back door address(0.0.0.0:8002)")
    flag.Usage = usage
    flag.Parse()

    if !options.gate {
        args := flag.Args()
        if len(args) < 1 {
            usage()
        } else {
            options.configFile = args[0]
        }
    }

    app := new(App)
    coor := NewCoor(options.gate, uint16(options.capacity))
    if options.gate {
        frontDoor := NewFrontServer(options.frontAddr, coor)
        backDoor  := NewBackServer(options.backAddr, coor)
        app.Add(frontDoor)
        app.Add(backDoor)
    } else {
        backClient := NewBackClient(options.configFile, options.backAddr, coor)
        app.Add(backClient)
    }


    err := app.Start()
    if err != nil {
        logger.Panicf("start gotunnel failed:%s", err.Error())
    }
    go handleSignal(app)

    app.Wait()
}

