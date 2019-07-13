package main

import (
	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
)

var region *ip2region.Ip2Region

func init() {
	region, _ = ip2region.New("/Users/fengma/Downloads/ip2region-master/data/ip2region.db")
}

func GetIpInfo(ip string) (ipInfo ip2region.IpInfo, err error) {
	return region.BtreeSearch(ip)
}
