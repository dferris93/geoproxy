package handler

import (
	"fmt"
	"sync"
)

type MockMutex struct {
	TryLockReturn bool
}

func (m *MockMutex) Lock() {}
func (m *MockMutex) Unlock() {}
func (m *MockMutex) TryLock() bool {
	return m.TryLockReturn
}

type GetCountryCodeMock struct {
	ReturnErr bool
	ReturnCountry string
	ReturnRegion string
	ReturnCached string
}

func (g *GetCountryCodeMock) GetCountryCode(ip string, m *sync.Mutex ) (string, string, string, error) {
	if g.ReturnErr {
		return "", "", "", nil
	} else {
		return g.ReturnCountry, g.ReturnRegion, g.ReturnCached, nil
	}
}


type MockCheckIP struct {
	CheckSubnetsReturn bool
	CheckIPTypeReturn int
	CheckIPTypeErr bool
}	

func (m *MockCheckIP) CheckSubnets(subnets []string, ip string) bool {
	return m.CheckSubnetsReturn
}

func (m *MockCheckIP) CheckIPType(ip string) (int, error) {
	if m.CheckIPTypeErr {
		return 0, fmt.Errorf("check ip type error")
	}
	return m.CheckIPTypeReturn, nil
}

