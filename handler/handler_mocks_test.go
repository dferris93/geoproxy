package handler

import "testing"

func TestHandlerMocks(t *testing.T) {
	mutex := &MockMutex{TryLockReturn: true}
	if !mutex.TryLock() {
		t.Fatal("expected TryLock true")
	}
	mutex.TryLockReturn = false
	if mutex.TryLock() {
		t.Fatal("expected TryLock false")
	}

	getter := &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA", ReturnCached: "cached"}
	country, region, cached, err := getter.GetCountryCode("1.2.3.4")
	if err != nil {
		t.Fatalf("GetCountryCode: %v", err)
	}
	if country != "US" || region != "CA" || cached != "cached" {
		t.Fatalf("unexpected response %q %q %q", country, region, cached)
	}

	getter.ReturnErr = true
	_, _, _, err = getter.GetCountryCode("1.2.3.4")
	if err != nil {
		t.Fatalf("expected nil error on ReturnErr path, got %v", err)
	}

	checker := &MockCheckIP{CheckIPTypeReturn: 6}
	if _, err := checker.CheckIPType("::1"); err != nil {
		t.Fatalf("CheckIPType: %v", err)
	}
	checker.CheckIPTypeErr = true
	if _, err := checker.CheckIPType("::1"); err == nil {
		t.Fatal("expected error on CheckIPTypeErr")
	}
}
