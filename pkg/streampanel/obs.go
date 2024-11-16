package streampanel

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/obs-grpc-proxy/protobuf/go/obs_grpc"
)

func (p *Panel) obsSetScene(
	ctx context.Context,
	sceneName string,
) error {
	obsServer, obsServerClose, err := p.StreamD.OBS(ctx)
	if obsServerClose != nil {
		defer obsServerClose()
	}
	if err != nil {
		return fmt.Errorf("unable to initialize a client to OBS: %w", err)
	}
	_, err = obsServer.SetCurrentProgramScene(ctx, &obs_grpc.SetCurrentProgramSceneRequest{
		SceneName: &sceneName,
	})
	if err != nil {
		return fmt.Errorf("unable to set the OBS scene: %w", err)
	}
	return nil
}
