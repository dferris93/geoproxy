package iptables

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockRunner struct {
	commands []string
	args     [][]string
	err      error
	callCh   chan struct{}
}

func (m *mockRunner) RunCommand(command string, args ...string) (string, error) {
	m.commands = append(m.commands, command)
	m.args = append(m.args, args)
	if m.callCh != nil {
		m.callCh <- struct{}{}
	}
	return "", m.err
}

type mockCheckIPs struct {
	ipType    int
	err       error
	errOnCall int
	calls     int
}

func (m *mockCheckIPs) CheckIPType(ip string) (int, error) {
	m.calls++
	if m.errOnCall > 0 && m.calls == m.errOnCall {
		return 0, m.err
	}
	return m.ipType, nil
}

func (m *mockCheckIPs) CheckSubnets(subnets []string, clientAddr string) bool {
	return false
}

func runBlockIPs(t *testing.T, ipt *IpTables, blockIPs chan string, ctx context.Context, done chan struct{}) {
	t.Helper()
	go func() {
		_ = ipt.BlockIPs(blockIPs, ctx)
		close(done)
	}()
}

func TestBlockIPsIPv4(t *testing.T) {
	blockIPs := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &mockRunner{callCh: make(chan struct{}, 1)}
	checker := &mockCheckIPs{ipType: 4}
	ipt := &IpTables{Chain: "INPUT", Action: "DROP", Runner: runner, CheckIPs: checker}
	done := make(chan struct{})

	runBlockIPs(t, ipt, blockIPs, ctx, done)
	blockIPs <- "1.2.3.4"
	select {
	case <-runner.callCh:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected iptables command to be executed")
	}
	cancel()
	<-done

	assert.Equal(t, []string{"iptables"}, runner.commands)
	if assert.Len(t, runner.args, 1) {
		assert.Equal(t, []string{"-A", "INPUT", "-s", "1.2.3.4", "-j", "DROP"}, runner.args[0])
	}
}

func TestBlockIPsIPv6(t *testing.T) {
	blockIPs := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &mockRunner{callCh: make(chan struct{}, 1)}
	checker := &mockCheckIPs{ipType: 6}
	ipt := &IpTables{Chain: "INPUT", Action: "DROP", Runner: runner, CheckIPs: checker}
	done := make(chan struct{})

	runBlockIPs(t, ipt, blockIPs, ctx, done)
	blockIPs <- "2001:db8::1"
	select {
	case <-runner.callCh:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected ip6tables command to be executed")
	}
	cancel()
	<-done

	assert.Equal(t, []string{"ip6tables"}, runner.commands)
	if assert.Len(t, runner.args, 1) {
		assert.Equal(t, []string{"-A", "INPUT", "-s", "2001:db8::1", "-j", "DROP"}, runner.args[0])
	}
}

func TestBlockIPsSkipsDuplicate(t *testing.T) {
	blockIPs := make(chan string, 3)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &mockRunner{callCh: make(chan struct{}, 2)}
	checker := &mockCheckIPs{ipType: 4}
	ipt := &IpTables{Chain: "INPUT", Action: "DROP", Runner: runner, CheckIPs: checker}
	done := make(chan struct{})

	runBlockIPs(t, ipt, blockIPs, ctx, done)
	blockIPs <- "1.2.3.4"
	blockIPs <- "1.2.3.4"
	blockIPs <- "2.2.2.2"
	for i := 0; i < 2; i++ {
		select {
		case <-runner.callCh:
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("expected iptables command to be executed")
		}
	}
	cancel()
	<-done

	assert.Equal(t, []string{"iptables", "iptables"}, runner.commands)
	if assert.Len(t, runner.args, 2) {
		assert.Equal(t, []string{"-A", "INPUT", "-s", "1.2.3.4", "-j", "DROP"}, runner.args[0])
		assert.Equal(t, []string{"-A", "INPUT", "-s", "2.2.2.2", "-j", "DROP"}, runner.args[1])
	}
}

func TestBlockIPsSkipsInvalidIP(t *testing.T) {
	blockIPs := make(chan string, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := &mockRunner{callCh: make(chan struct{}, 1)}
	checker := &mockCheckIPs{ipType: 4, err: errors.New("bad ip"), errOnCall: 1}
	ipt := &IpTables{Chain: "INPUT", Action: "DROP", Runner: runner, CheckIPs: checker}
	done := make(chan struct{})

	runBlockIPs(t, ipt, blockIPs, ctx, done)
	blockIPs <- "bad-ip"
	blockIPs <- "1.2.3.4"
	select {
	case <-runner.callCh:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected iptables command to be executed")
	}
	cancel()
	<-done

	assert.Equal(t, []string{"iptables"}, runner.commands)
	if assert.Len(t, runner.args, 1) {
		assert.Equal(t, []string{"-A", "INPUT", "-s", "1.2.3.4", "-j", "DROP"}, runner.args[0])
	}
}
