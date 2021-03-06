package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/vonaka/shreplic/client/base"
	"github.com/vonaka/shreplic/curp"
	"github.com/vonaka/shreplic/paxoi"
	"github.com/vonaka/shreplic/tools/dlog"
)

var (
	clientId       = flag.String("id", "", "The id of the client. Default is RFC 4122 nodeID")
	maddr          = flag.String("maddr", "", "Master address")
	mport          = flag.Int("mport", 7087, "Master port")
	reqNum         = flag.Int("q", 1000, "Total number of requests")
	writes         = flag.Int("w", 50, "Percentage of updates (writes)")
	psize          = flag.Int("psize", 100, "Payload size for writes")
	noLeader       = flag.Bool("e", false, "Egalitarian (no leader)")
	fast           = flag.Bool("f", false, "Send message directly to all replicas")
	lread          = flag.Bool("l", false, "Execute reads at the closest (local) replica")
	conflicts      = flag.Int("c", 0, "Percentage of conflicts")
	verbose        = flag.Bool("v", false, "Verbose mode")
	collocatedWith = flag.String("server", "", "Server with which this client is collocated")
	myAddr         = flag.String("addr", "", "Client address (this machine)")
	cloneNb        = flag.Int("clone", 0, "Number of clones (unique clients acting like this one)")
	logFile        = flag.String("logf", "", "Path to the log file")
	paxoiClient    = flag.Bool("paxoi", false, "Run Paxoi external client")
	curpClient     = flag.Bool("curp", false, "Run CURP external client")
	args           = flag.String("args", "", "Custom arguments")
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	for i := 0; i < *cloneNb+1; i++ {
		wg.Add(1)
		go func(i int) {
			runSimpleClient(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func runSimpleClient(i int) {
	var l *log.Logger
	if i == 0 {
		l = newLogger(*logFile)
	} else {
		l = newLogger(*logFile + strconv.Itoa(i))
	}
	if *paxoiClient {
		c := paxoi.NewClient(*maddr, *collocatedWith, *mport, *reqNum, *writes,
			*psize, *conflicts, *fast, *lread, *noLeader, *verbose, l, *args)
		if c == nil {
			return
		}
		err := c.Run()
		if err != nil {
			fmt.Println(err)
		}
	} else if *curpClient {
		c := curp.NewClient(*maddr, *collocatedWith, *mport, *reqNum, *writes,
			*psize, *conflicts, *fast, *lread, *noLeader, *verbose, l, *args)
		if c == nil {
			return
		}
		err := c.Run()
		if err != nil {
			fmt.Println(err)
		}
	} else {
		c := base.NewSimpleClient(*maddr, *collocatedWith, *mport, *reqNum,
			*writes, *psize, *conflicts, *fast, *lread, *noLeader, *verbose, l)
		err := c.Run()
		for err != nil {
			if err == io.EOF {
				time.Sleep(time.Second)
				err = c.Rerun()
			} else {
				fmt.Println(err)
				return
			}
		}
	}
}

func newLogger(logPath string) *log.Logger {
	logF := os.Stdout
	if logPath != "" {
		var err error
		logF, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("Can't open log file:", logPath)
		}
	}
	return dlog.NewFileLogger(logF)
}
