package main

import (
	"context"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/projecteru2/core/migration/etcd"
	"github.com/projecteru2/core/store/etcdv3"
	"github.com/projecteru2/core/types"
	"github.com/projecteru2/core/utils"
	"github.com/projecteru2/core/versioninfo"
	"github.com/sanity-io/litter"
	log "github.com/sirupsen/logrus"
)

var configPath string

func serve() {
	if configPath == "" {
		log.Fatal("[main] Config path must be set")
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("[main] %v", err)
	}
	logLevel := "INFO"
	if config.LogLevel != "" {
		logLevel = config.LogLevel
	}
	if err := setupLog(logLevel); err != nil {
		log.Fatalf("[main] %v", err)
	}

	oldStore, _ := etcdstore.New(config) // TODO error process
	newStore, _ := etcdv3.New(config)

	migration(config, oldStore, newStore)
}

func migration(config types.Config, old *etcdstore.Krypton, new *etcdv3.Mercury) {
	ctx := context.Background()
	pods, _ := old.GetAllPods(ctx)
	litter.Dump(pods)
	for _, pod := range pods {
		p, _ := new.AddPod(ctx, pod.Name, pod.Favor, pod.Desc) // TODO error process
		litter.Dump(p)
		nodes, _ := old.GetNodesByPod(ctx, pod.Name)
		for _, node := range nodes {
			ca, certs, key := old.GetNodeCerts(ctx, p.Name, node.Name)
			n, err := new.AddNode(ctx, node.Name, node.Endpoint, p.Name, ca, certs, key, len(node.CPU), config.Scheduler.ShareBase, node.MemCap, node.Labels)
			if err != nil {
				continue
			}
			// should update node resource
			n.CPU = node.CPU
			new.UpdateNode(ctx, n)
		}
	}

	containers, _ := old.ListContainers(ctx, "", "", "")
	for _, container := range containers {
		new.AddContainer(ctx, container)
	}
	log.Info("Done")
}

func setupLog(l string) error {
	level, err := log.ParseLevel(l)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	formatter := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	return nil
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(versioninfo.VersionString())
	}

	app := cli.NewApp()
	app.Name = versioninfo.NAME
	app.Usage = "Run eru core"
	app.Version = versioninfo.VERSION
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Value:       "/etc/eru/core.yaml",
			Usage:       "config file path for core, in yaml",
			Destination: &configPath,
			EnvVar:      "ERU_CONFIG_PATH",
		},
	}
	app.Action = func(c *cli.Context) error {
		serve()
		return nil
	}

	app.Run(os.Args)
}
