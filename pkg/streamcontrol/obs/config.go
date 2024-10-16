package obs

import (
	streamctl "github.com/xaionaro-go/streamctl/pkg/streamcontrol"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol/obs/types"
)

const ID = types.ID

type Config = types.Config
type SceneName = types.SceneName
type SceneRule = types.SceneRule
type SceneRules = types.SceneRules
type StreamProfile = types.StreamProfile
type PlatformSpecificConfig = types.PlatformSpecificConfig

func InitConfig(cfg streamctl.Config) {
	types.InitConfig(cfg)
}
