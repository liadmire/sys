package sys

import (
	"fmt"
	"math/big"
	"net"
)

// StringIPToInt x
func StringIPToInt(ipaddr string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ipaddr).To4())
	return ret.Int64()
}

// IntIPToString x
func IntIPToString(ipaddr int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ipaddr>>24), byte(ipaddr>>16), byte(ipaddr>>8), byte(ipaddr))
}
