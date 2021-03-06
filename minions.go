package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	"github.com/projectcalico/libcalico-go/lib/clientv3"
	"github.com/projecteru2/minions/driver"
	"github.com/projecteru2/minions/versioninfo"
	log "github.com/sirupsen/logrus"
)

var (
	cnmName  string
	ipamName string
	debug    bool

	config *apiconfig.CalicoAPIConfig
	client clientv3.Interface
)

func initializeClient() {
	var err error

	if config, err = apiconfig.LoadClientConfig(""); err != nil {
		panic(err)
	}
	if client, err = clientv3.New(*config); err != nil {
		panic(err)
	}

	if debug {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Debug logging enabled")
	}
}

func serve() {
	initializeClient()

	errChannel := make(chan error)
	networkHandler := network.NewHandler(driver.NewNetworkDriver(client))
	ipamHandler := ipam.NewHandler(driver.NewIpamDriver(client))

	go func(c chan error) {
		log.Infoln("calico-net has started.")
		err := networkHandler.ServeUnix(cnmName, 0)
		log.Infoln("calico-net has stopped working.")
		c <- err
	}(errChannel)

	go func(c chan error) {
		log.Infoln("calico-ipam has started.")
		err := ipamHandler.ServeUnix(ipamName, 0)
		log.Infoln("calico-ipam has stopped working.")
		c <- err
	}(errChannel)

	err := <-errChannel

	log.Fatalln(err)
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(versioninfo.VersionString())
	}

	app := cli.NewApp()
	app.Name = versioninfo.NAME
	app.Usage = "Run eru minions"
	app.Version = versioninfo.VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "cnm",
			Value:       "calico",
			Usage:       "CNM name",
			Destination: &cnmName,
			EnvVar:      "CALICO_CNM",
		},
		cli.StringFlag{
			Name:        "ipam",
			Value:       "calico-ipam",
			Usage:       "ipam name",
			Destination: &ipamName,
			EnvVar:      "CALICO_IPAM",
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug or not",
			Destination: &debug,
			EnvVar:      "CALICO_DEBUG",
		},
	}
	app.Action = func(c *cli.Context) error {
		serve()
		return nil
	}

	app.Run(os.Args)
}
