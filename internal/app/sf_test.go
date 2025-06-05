package app

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"net"
	"sync"
	"testing"
)

// 1930212796655771648
func TestInitServices(t *testing.T) {
	t.Log(GenerateUserID())
}
