package handler

import (
	"crypto/sha1"
	"encoding/binary"
	"github.com/bwmarrin/snowflake"
	"net"
	"sync"
)

var (
	node *snowflake.Node
	once sync.Once
)

// 自动基于 IP 派生 nodeID
func getNodeIDFromIP() int64 {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			h := sha1.Sum([]byte(ipnet.IP.String()))
			nodeID := binary.BigEndian.Uint16(h[:2]) % 1024 // 取前两字节并限制范围
			return int64(nodeID)
		}
	}
	return 1 // fallback 默认 nodeID
}

// InitSnowflakeNode 初始化节点
func InitSnowflakeNode() {
	once.Do(func() {
		nodeID := getNodeIDFromIP()
		var err error
		node, err = snowflake.NewNode(nodeID)
		if err != nil {
			panic(err)
		}
	})
}

// GenerateUserID 生成 Snowflake ID
func GenerateUserID() string {
	InitSnowflakeNode()
	return node.Generate().Base58()
}
