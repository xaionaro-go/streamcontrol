package main

import (
	"bytes"
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/streamctl/pkg/mainprocess"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol"
	"github.com/xaionaro-go/streamctl/pkg/streamd/grpc/go/streamd_grpc"
	"github.com/xaionaro-go/streamctl/pkg/streampanel"
	_ "github.com/xaionaro-go/streamctl/pkg/streamserver"
)

const forceNetPProfOnAndroid = true

func main() {
	flags := parseFlags()
	ctx := getContext(flags)
	defer belt.Flush(ctx)
	cancelFunc := initRuntime(ctx, flags, ProcessNameMain)
	defer cancelFunc()

	if flags.Subprocess != "" {
		runSubprocess(ctx, flags.Subprocess)
		return
	}

	if flags.SplitProcess {
		runSplitProcesses(ctx, flags)
		return
	}

	runPanel(ctx, flags, nil)
}

func runPanel(
	ctx context.Context,
	flags Flags,
	mainProcess *mainprocess.Client,
) {
	logger.Debugf(ctx, "runPanel: %#+v", flags)
	defer logger.Debugf(ctx, "/runPanel")

	var opts []streampanel.Option
	if flags.RemoteAddr != "" {
		opts = append(opts, streampanel.OptionRemoteStreamDAddr(flags.RemoteAddr))
	}

	panel, panelErr := streampanel.New(flags.ConfigPath, opts...)
	if panelErr != nil {
		logger.Fatal(ctx, panelErr)
	}

	if !flags.SplitProcess && flags.ListenAddr != "" {
		listener, grpcServer, streamdGRPC := initGRPCServer(ctx, panel.StreamD, flags.ListenAddr)

		// to erase an oauth request answered locally from "UnansweredOAuthRequests" in the GRPC server:
		panel.OnInternallySubmittedOAuthCode = func(
			ctx context.Context,
			platID streamcontrol.PlatformName,
			code string,
		) error {
			_, err := streamdGRPC.SubmitOAuthCode(ctx, &streamd_grpc.SubmitOAuthCodeRequest{
				PlatID: string(platID),
				Code:   code,
			})
			return err
		}

		err := grpcServer.Serve(listener)
		if err != nil {
			logger.Fatalf(ctx, "unable to server the gRPC server: %v", err)
		}
	}

	if mainProcess != nil {
		go func() {
			err := mainProcess.Serve(
				ctx,
				func(ctx context.Context, source mainprocess.ProcessName, content any) error {
					switch msg := content.(type) {
					case StreamDDied:
						logger.Errorf(ctx, "streamd died, killing myself as well (to get reborn)")
						os.Exit(0)
					case UpdateStreamDConfig:
						_, err := panel.Config.BuiltinStreamD.ReadFrom(bytes.NewReader([]byte(msg.Config)))
						if err != nil {
							err := fmt.Errorf("unable to deserialize the updated streamd config: %w", err)
							logger.Errorf(ctx, "%s", err)
							return err
						}
						err = panel.SaveConfig(ctx)
						if err != nil {
							err := fmt.Errorf("unable to save the updated streamd config: %w", err)
							logger.Errorf(ctx, "%s", err)
							return err
						}
					}
					return nil
				},
			)
			logger.Fatalf(ctx, "communication (with the main process) error: %v", err)
		}()
	}

	var loopOpts []streampanel.LoopOption
	if flags.Page != "" {
		loopOpts = append(loopOpts, streampanel.LoopOptionStartingPage(flags.Page))
	}
	err := panel.Loop(ctx, loopOpts...)
	if err != nil {
		logger.Fatal(ctx, err)
	}
}
