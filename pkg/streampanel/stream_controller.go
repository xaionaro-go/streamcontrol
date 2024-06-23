package streampanel

import (
	"context"
	"errors"
	"fmt"

	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol/obs"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol/twitch"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol/youtube"
)

var ErrSkipBackend = errors.New("backend was skipped")

func newOBS(
	ctx context.Context,
	cfg *streamcontrol.AbstractPlatformConfig,
	setConnectionInfo func(context.Context, *streamcontrol.PlatformConfig[obs.PlatformSpecificConfig, obs.StreamProfile]) (bool, error),
	saveCfgFunc func(*streamcontrol.AbstractPlatformConfig) error,
) (
	*obs.OBS,
	error,
) {
	platCfg := streamcontrol.ConvertPlatformConfig[obs.PlatformSpecificConfig, obs.StreamProfile](
		ctx, cfg,
	)
	if platCfg == nil {
		return nil, fmt.Errorf("OBS config was not found")
	}

	if cfg.Enable != nil && !*cfg.Enable {
		return nil, ErrSkipBackend
	}

	hadSetNewConnectionInfo := false
	if platCfg.Config.Host == "" || platCfg.Config.Port == 0 {
		ok, err := setConnectionInfo(ctx, platCfg)
		if !ok {
			err := saveCfgFunc(&streamcontrol.AbstractPlatformConfig{
				Enable:         platCfg.Enable,
				Config:         platCfg.Config,
				StreamProfiles: streamcontrol.ToAbstractStreamProfiles(platCfg.StreamProfiles),
			})
			if err != nil {
				logger.Error(ctx, err)
			}
			return nil, ErrSkipBackend
		}
		if err != nil {
			return nil, fmt.Errorf("unable to set connection info: %w", err)
		}
		hadSetNewConnectionInfo = true
	}

	logger.Debugf(ctx, "OBS config: %#+v", platCfg)
	cfg = streamcontrol.ToAbstractPlatformConfig(ctx, platCfg)
	obs, err := obs.New(ctx, *platCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize OBS client: %w", err)
	}
	if hadSetNewConnectionInfo {
		logger.Debugf(ctx, "confirmed new OBS connection info, saving it")
		if err := saveCfgFunc(cfg); err != nil {
			return nil, fmt.Errorf("unable to save the configuration: %w", err)
		}
	}
	return obs, nil

}

func newTwitch(
	ctx context.Context,
	cfg *streamcontrol.AbstractPlatformConfig,
	setUserData func(context.Context, *streamcontrol.PlatformConfig[twitch.PlatformSpecificConfig, twitch.StreamProfile]) (bool, error),
	saveCfgFunc func(*streamcontrol.AbstractPlatformConfig) error,
	customOAuthHandler twitch.OAuthHandler,
) (
	*twitch.Twitch,
	error,
) {
	platCfg := streamcontrol.ConvertPlatformConfig[twitch.PlatformSpecificConfig, twitch.StreamProfile](
		ctx, cfg,
	)
	if platCfg == nil {
		return nil, fmt.Errorf("twitch config was not found")
	}

	if cfg.Enable != nil && !*cfg.Enable {
		return nil, ErrSkipBackend
	}

	hadSetNewUserData := false
	if platCfg.Config.Channel == "" || platCfg.Config.ClientID == "" || platCfg.Config.ClientSecret == "" {
		ok, err := setUserData(ctx, platCfg)
		if !ok {
			err := saveCfgFunc(&streamcontrol.AbstractPlatformConfig{
				Enable:         platCfg.Enable,
				Config:         platCfg.Config,
				StreamProfiles: streamcontrol.ToAbstractStreamProfiles(platCfg.StreamProfiles),
			})
			if err != nil {
				logger.Error(ctx, err)
			}
			return nil, ErrSkipBackend
		}
		if err != nil {
			return nil, fmt.Errorf("unable to set user info: %w", err)
		}
		hadSetNewUserData = true
	}

	logger.Debugf(ctx, "twitch config: %#+v", platCfg)
	platCfg.Config.CustomOAuthHandler = customOAuthHandler
	cfg = streamcontrol.ToAbstractPlatformConfig(ctx, platCfg)
	twitch, err := twitch.New(ctx, *platCfg,
		func(c twitch.Config) error {
			return saveCfgFunc(&streamcontrol.AbstractPlatformConfig{
				Enable:         c.Enable,
				Config:         c.Config,
				StreamProfiles: streamcontrol.ToAbstractStreamProfiles(c.StreamProfiles),
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Twitch client: %w", err)
	}
	if hadSetNewUserData {
		logger.Debugf(ctx, "confirmed new twitch user data, saving it")
		if err := saveCfgFunc(cfg); err != nil {
			return nil, fmt.Errorf("unable to save the configuration: %w", err)
		}
	}
	return twitch, nil
}

func newYouTube(
	ctx context.Context,
	cfg *streamcontrol.AbstractPlatformConfig,
	setUserData func(context.Context, *streamcontrol.PlatformConfig[youtube.PlatformSpecificConfig, youtube.StreamProfile]) (bool, error),
	saveCfgFunc func(*streamcontrol.AbstractPlatformConfig) error,
	customOAuthHandler youtube.OAuthHandler,
) (
	*youtube.YouTube,
	error,
) {
	platCfg := streamcontrol.ConvertPlatformConfig[youtube.PlatformSpecificConfig, youtube.StreamProfile](
		ctx, cfg,
	)
	if platCfg == nil {
		return nil, fmt.Errorf("youtube config was not found")
	}

	if cfg.Enable != nil && !*cfg.Enable {
		return nil, ErrSkipBackend
	}

	hadSetNewUserData := false
	if platCfg.Config.ClientID == "" || platCfg.Config.ClientSecret == "" {
		ok, err := setUserData(ctx, platCfg)
		if !ok {
			err := saveCfgFunc(&streamcontrol.AbstractPlatformConfig{
				Enable:         platCfg.Enable,
				Config:         platCfg.Config,
				StreamProfiles: streamcontrol.ToAbstractStreamProfiles(platCfg.StreamProfiles),
			})
			if err != nil {
				logger.Error(ctx, err)
			}
			return nil, ErrSkipBackend
		}
		if err != nil {
			return nil, fmt.Errorf("unable to set user info: %w", err)
		}
		hadSetNewUserData = true
	}

	logger.Debugf(ctx, "youtube config: %#+v", platCfg)
	platCfg.Config.CustomOAuthHandler = customOAuthHandler
	cfg = streamcontrol.ToAbstractPlatformConfig(ctx, platCfg)
	yt, err := youtube.New(ctx, *platCfg,
		func(c youtube.Config) error {
			return saveCfgFunc(&streamcontrol.AbstractPlatformConfig{
				Enable:         c.Enable,
				Config:         c.Config,
				StreamProfiles: streamcontrol.ToAbstractStreamProfiles(c.StreamProfiles),
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize YouTube client: %w", err)
	}
	if hadSetNewUserData {
		logger.Debugf(ctx, "confirmed new youtube user data, saving it")
		if err := saveCfgFunc(cfg); err != nil {
			return nil, fmt.Errorf("unable to save the configuration: %w", err)
		}
	}
	return yt, nil
}

func (p *Panel) initOBSBackend(ctx context.Context) error {
	obs, err := newOBS(
		ctx,
		p.data.Backends[obs.ID],
		p.inputOBSConnectInfo,
		func(cfg *streamcontrol.AbstractPlatformConfig) error {
			return p.savePlatformConfig(ctx, obs.ID, cfg)
		},
	)
	if err != nil {
		return err
	}
	p.streamControllers.OBS = obs
	return nil
}

func (p *Panel) initTwitchBackend(ctx context.Context) error {
	twitch, err := newTwitch(
		ctx,
		p.data.Backends[twitch.ID],
		p.inputTwitchUserInfo,
		func(cfg *streamcontrol.AbstractPlatformConfig) error {
			return p.savePlatformConfig(ctx, twitch.ID, cfg)
		},
		p.oauthHandlerTwitch)
	if err != nil {
		return err
	}
	p.streamControllers.Twitch = twitch
	return nil
}

func (p *Panel) initYouTubeBackend(ctx context.Context) error {
	youTube, err := newYouTube(
		ctx,
		p.data.Backends[youtube.ID],
		p.inputYouTubeUserInfo,
		func(cfg *streamcontrol.AbstractPlatformConfig) error {
			return p.savePlatformConfig(ctx, youtube.ID, cfg)
		},
		p.oauthHandlerYouTube,
	)
	if err != nil {
		return err
	}
	p.streamControllers.YouTube = youTube
	return nil
}
