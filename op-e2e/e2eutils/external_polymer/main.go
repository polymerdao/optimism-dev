package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/external"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Execute based on the config in this file")
	flag.Parse()
	if err := run(configPath); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

// relative to where the shim binary lives
const opPolymerBin = "polymer-peptide"

func run(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("must supply a '--config <path>' flag")
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("could not open config: %w", err)
	}

	var config external.Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return fmt.Errorf("could not decode config file: %w", err)
	}

	binPath, err := filepath.Abs(opPolymerBin)
	if err != nil {
		return fmt.Errorf("could not get absolute path of op-polymer")
	}
	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("could not locate op-polymer in working directory")
	}

	fmt.Printf("================== op-polymer shim initializing chain config ==========================\n")
	if err := initialize(binPath, config); err != nil {
		return fmt.Errorf("could not initialize datadir: %s %w", binPath, err)
	}

	fmt.Printf("================== op-polymer shim sealing chain ==========================\n")
	if err := seal(binPath, config); err != nil {
		return fmt.Errorf("could not seal datadir: %s %w", binPath, err)
	}

	sess, err := execute(binPath, config)
	if err != nil {
		return fmt.Errorf("could not execute polymer: %w", err)
	}
	defer sess.Close()

	fmt.Printf("==================    op-polymer shim encoding ready-file  ==========================\n")
	if err := external.AtomicEncode(config.EndpointsReadyPath, sess.endpoints); err != nil {
		return fmt.Errorf("could not encode endpoints")
	}

	fmt.Printf("==================    op-polymer shim awaiting termination  ==========================\n")

	// Listen for kill signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-sigCh
	sess.Close()

	select {
	case <-sess.session.Exited:
		return fmt.Errorf("geth exited")
	case <-time.After(30 * time.Minute):
		return fmt.Errorf("exiting after 30 minute timeout")
	}
}

func initialize(binPath string, config external.Config) error {
	cmd := exec.Command(
		binPath,
		"init",
		"--home", config.DataDir,
		"--l1-hash", config.L1Hash,
		"--l1-height", fmt.Sprintf("%d", config.L1Height),
	)
	return cmd.Run()
}

func seal(binPath string, config external.Config) error {
	var outb, errb bytes.Buffer
	cmd := exec.Command(
		binPath,
		"seal",
		"--home", config.DataDir,
		"--genesis-time", fmt.Sprintf("%d", config.L2Time),
	)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	fmt.Printf("%v\n", outb.String())
	fmt.Printf("%v\n", errb.String())
	return err
}

type gethSession struct {
	session   *gexec.Session
	endpoints *external.Endpoints
}

func (es *gethSession) Close() {
	es.session.Terminate()
	select {
	case <-time.After(5 * time.Second):
		es.session.Kill()
	case <-es.session.Exited:
	}
}

func execute(binPath string, config external.Config) (*gethSession, error) {
	cmd := exec.Command(binPath, "start",
		"--app-rpc-address", "localhost:0",
		"--ee-http-server-address", "localhost:0",
		"--home", config.DataDir,
	)
	sess, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("could not start op-polymer session: %w", err)
	}
	matcher := gbytes.Say("Execution engine rpc server enabled")
	var httpUrl, wsUrl string
	urlRE := regexp.MustCompile(`Execution engine rpc server enabled\s+http=(.+)\sws=(.+)`)
	for httpUrl == "" && wsUrl == "" {
		match, err := matcher.Match(sess.Out)
		if err != nil {
			return nil, fmt.Errorf("could not execute matcher")
		}
		if !match {
			if sess.Out.Closed() {
				return nil, fmt.Errorf("op-polymer exited before announcing http ports")
			}
			// Wait for a bit more output, then try again
			time.Sleep(10 * time.Millisecond)
			continue
		}

		for _, line := range strings.Split(string(sess.Out.Contents()), "\n") {
			found := urlRE.FindStringSubmatch(line)
			if len(found) == 3 {
				httpUrl = found[1]
				wsUrl = found[2]
				continue
			}
		}
	}

	genesisFile, err := os.Open(filepath.Join(config.DataDir, "config", "genesis.json"))
	if err != nil {
		return nil, fmt.Errorf("could not open genesis file: %w", err)
	}

	var genesis struct {
		GenesisBlock eth.BlockID `json:"genesis_block"`
	}
	if err := json.NewDecoder(genesisFile).Decode(&genesis); err != nil {
		return nil, fmt.Errorf("could not decode genesis file: %w", err)
	}

	return &gethSession{
		session: sess,
		endpoints: &external.Endpoints{
			HTTPEndpoint:       httpUrl,
			WSEndpoint:         wsUrl,
			HTTPAuthEndpoint:   httpUrl,
			WSAuthEndpoint:     wsUrl,
			GenesisBlockHash:   genesis.GenesisBlock.Hash.Hex(),
			GenesisBlockHeight: strconv.FormatUint(genesis.GenesisBlock.Number, 10),
		},
	}, nil
}
