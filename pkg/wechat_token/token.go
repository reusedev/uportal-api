package wechat_token

import (
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/robfig/cron"
	"sync"
)

var (
	token string
	lock  = new(sync.RWMutex)
)

// TokenJob 定时刷新 token
func TokenJob() {
	refreshToken()
	c := cron.New()
	c.AddFunc("@every 5m", refreshToken)
	c.Start()
}

func GetToken() string {
	lock.RLock()
	defer lock.RUnlock()
	return token
}

func refreshToken() {
	t := getToken(config.GlobalConfig.Wechat.MiniProgram.AppID, config.GlobalConfig.Wechat.MiniProgram.AppSecret)
	if t != "" {
		lock.Lock()
		token = t
		lock.Unlock()
	}
}
