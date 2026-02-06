package handler

import (
	"context"
	"fmt"
)

type GetCountryCodeMock struct {
	ReturnErr     bool
	ReturnCountry string
	ReturnRegion  string
	ReturnCached  string
}

func (g *GetCountryCodeMock) GetCountryCode(_ context.Context, ip string) (string, string, string, error) {
	if g.ReturnErr {
		return "", "", "", nil
	} else {
		return g.ReturnCountry, g.ReturnRegion, g.ReturnCached, nil
	}
}

type MockCheckIP struct {
	CheckSubnetsReturn bool
	CheckIPTypeReturn  int
	CheckIPTypeErr     bool
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
