package main

import (
	"fmt"
	"os"
	"time"

	"github.com/facebookincubator/go-belt/tool/logger"
	flag "github.com/xaionaro-go/streamctl/pkg/ffflag"
)

type Flags struct {
	Inputs                []Resource
	ListenControlSocket   string
	ListenNetPprof        string
	LoggerLevel           logger.Level
	LogstashAddr          string
	SentryDSN             string
	LogFile               string
	LockTimeout           time.Duration
	InsecureDebug         bool
	RemoveSecretsFromLogs bool
	VideoEncoder          Encoder
	AudioEncoder          Encoder
	Output                Resource
}

type Encoder struct {
	Codec   string
	Options []string
}

type Resource struct {
	URL     string
	Options []string
}

func parseFlags(args []string) Flags {
	p := flag.NewParser()
	inputsFlag := flag.AddParameter(p, "i", true, ptr(flag.StringsAsSeparateFlags(nil)))
	encoderBothFlag := flag.AddParameter(p, "c", true, ptr(flag.String("")))
	encoderVideoFlag := flag.AddParameter(p, "c:v", true, ptr(flag.String("")))
	encoderAudioFlag := flag.AddParameter(p, "c:a", true, ptr(flag.String("")))
	listenControlSocket := flag.AddParameter(p, "listen_control", false, ptr(flag.String("")))
	listenNetPprof := flag.AddParameter(p, "listen_net_pprof", false, ptr(flag.String("")))
	loggerLevel := flag.AddParameter(p, "v", false, ptr(flag.LogLevel(logger.LevelInfo)))
	logstashAddr := flag.AddParameter(p, "logstash_addr", false, ptr(flag.String("")))
	sentryDSN := flag.AddParameter(p, "sentry_dsn", false, ptr(flag.String("")))
	logFile := flag.AddParameter(p, "log_file", false, ptr(flag.String("")))
	lockTimeout := flag.AddParameter(p, "lock_timeout", false, ptr(flag.Duration(time.Minute)))
	insecureDebug := flag.AddParameter(p, "insecure_debug", false, ptr(flag.Bool(false)))
	removeSecretsFromLogs := flag.AddParameter(p, "remove_secrets_from_logs", false, ptr(flag.Bool(false)))
	version := flag.AddFlag(p, "version", false)

	err := p.Parse(args)
	assertNoError(err)

	if len(p.CollectedNonFlags) != 1 {
		panic(fmt.Errorf("expected exactly one output, but received %d", len(p.CollectedNonFlags)))
	}
	output := Resource{
		URL:     p.CollectedNonFlags[0],
		Options: p.CollectedUnknownOptions,
	}

	var inputs []Resource
	for idx, input := range inputsFlag.Value() {
		collectedOptions := inputsFlag.CollectedUnknownOptions[idx]
		inputs = append(inputs, Resource{
			URL:     input,
			Options: collectedOptions,
		})
	}

	if version.Value() {
		printBuildInfo(os.Stdout)
		os.Exit(0)
	}

	flags := Flags{
		ListenControlSocket: listenControlSocket.Value(),
		ListenNetPprof:      listenNetPprof.Value(),
		LoggerLevel:         loggerLevel.Value(),
		LogstashAddr:        logstashAddr.Value(),
		SentryDSN:           sentryDSN.Value(),
		LogFile:             logFile.Value(),
		LockTimeout:         lockTimeout.Value(),

		InsecureDebug:         insecureDebug.Value(),
		RemoveSecretsFromLogs: removeSecretsFromLogs.Value(),

		Inputs: inputs,
		Output: output,
	}

	if v := encoderBothFlag.Value(); v != "" {
		flags.AudioEncoder = Encoder{
			Codec:   v,
			Options: encoderBothFlag.CollectedUnknownOptions[0],
		}
		flags.VideoEncoder = Encoder{
			Codec:   v,
			Options: encoderBothFlag.CollectedUnknownOptions[0],
		}
	}

	if v := encoderVideoFlag.Value(); v != "" {
		flags.VideoEncoder = Encoder{
			Codec:   v,
			Options: encoderVideoFlag.CollectedUnknownOptions[0],
		}
	}

	if v := encoderAudioFlag.Value(); v != "" {
		flags.AudioEncoder = Encoder{
			Codec:   v,
			Options: encoderAudioFlag.CollectedUnknownOptions[0],
		}
	}

	return flags
}
