package wechat_token

import (
	"fmt"
	"sync"

	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/logs"
	"github.com/robfig/cron"
)

var (
	token string
	lock  = new(sync.RWMutex)
)

// TokenJob 定时刷新 token
func TokenJob() {
	refreshToken()
	c := cron.New()
	c.AddFunc("@every 4m", refreshToken)
	c.Start()
}

func GetToken() string {
	lock.RLock()
	defer lock.RUnlock()
	return token
}

func refreshToken() {
	t, err := getStableAccessToken(config.GlobalConfig.Wechat.MiniProgram.AppID, config.GlobalConfig.Wechat.MiniProgram.AppSecret, false)
	if err != nil {
		//todo: 报警
		logs.Business().Error(fmt.Sprintf("get stable access token failed: %s", err.Error()))
		return
	}
	if t != "" {
		lock.Lock()
		token = t
		lock.Unlock()
	}
}
