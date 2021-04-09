// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package genesis

import (
	"math/big"
	"net"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func splitSubnet(base *net.IPNet, newBits int, logger *logrus.Entry) ([]net.IPNet, error) {
	ip := base.IP
	mask := base.Mask

	parentLen, addrLen := mask.Size()
	newPrefixLen := parentLen + newBits

	if newPrefixLen > addrLen {
		return nil, errors.Errorf("insufficient address space to extend prefix of %d by %d", parentLen, newBits)
	}

	maxNetNum := uint64(1<<uint64(newBits)) - 1

	logger.Infof("Parent subnet %s will be split into %d subnets", string(ip), maxNetNum)
	var i uint64
	var subnets []net.IPNet
	for i = 0; i <= maxNetNum; i++ {
		subnets = append(subnets, net.IPNet{
			IP:   insertNumIntoIP(ip, new(big.Int).SetUint64(i), newPrefixLen),
			Mask: net.CIDRMask(newPrefixLen, addrLen),
		})
	}

	return subnets, nil
}

func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		return val, 32
	} else if len(ip) == net.IPv6len {
		return val, 128
	} else {
		panic(errors.Errorf("Unsupported address length %d", len(ip)))
	}
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}

func insertNumIntoIP(ip net.IP, bigNum *big.Int, prefixLen int) net.IP {
	ipInt, totalBits := ipToInt(ip)
	bigNum.Lsh(bigNum, uint(totalBits-prefixLen))
	ipInt.Or(ipInt, bigNum)
	return intToIP(ipInt, totalBits)
}
