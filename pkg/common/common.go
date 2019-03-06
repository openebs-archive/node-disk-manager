package common

import (
	"fmt"
	"sync"
	"time"
)

type NodeLastTime struct {
	LastTimeStamp time.Time
}

type NodeLive struct {
	Mu      sync.Mutex // Gaurd NodeMap
	NodeMap map[string]NodeLastTime
}

var NodeLiveInfo *NodeLive

func InitNodeMap() {
	NodeLiveInfo = &NodeLive{}
	NodeLiveInfo.NodeMap = make(map[string]NodeLastTime)
}

func UpdateNodeLivenessTimeStamp(hostName string, curr_time time.Time) {
	fmt.Println("Updating NodeMap for host:", hostName)
	NodeLiveInfo.Mu.Lock()
	NodeLiveInfo.NodeMap[hostName] = NodeLastTime{LastTimeStamp: curr_time}
	NodeLiveInfo.Mu.Unlock()
	return
}

func FindInactiveNodes(nodeExpiryTimeInterval time.Duration, NodeList []string) []string {
	var InactiveNodeList []string
	curr := time.Now()

	if len(NodeList) != 0 {
		for _, host := range NodeList {
			_, ok := NodeLiveInfo.NodeMap[host]
			if !ok {
				fmt.Println("Node not found in NodeMap:", host)
				InactiveNodeList = append(InactiveNodeList, host)
			}
		}

		if len(InactiveNodeList) != 0 {
			fmt.Println("InactiveNodeList:", InactiveNodeList)
		}
	}

	NodeLiveInfo.Mu.Lock()
	for k := range NodeLiveInfo.NodeMap {
		if NodeLiveInfo.NodeMap[k].LastTimeStamp.Add(nodeExpiryTimeInterval).Before(curr) {
			fmt.Println("Inactive Node:", k, "TimeStamp:", NodeLiveInfo.NodeMap[k])
			InactiveNodeList = append(InactiveNodeList, k)
		}
	}
	NodeLiveInfo.Mu.Unlock()
	return InactiveNodeList
}
