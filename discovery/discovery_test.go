package discovery_test

import (
	"fmt"
	discovery "lan-discovery/discovery"
	"testing"
	"time"
)

func TestDiscovery(t *testing.T) {
	errChan := make(chan error)
	doneChan := make(chan struct{})
	go func() {
		time.Sleep(1000)
		c, err := discovery.Discover()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println(c)
		doneChan <- struct{}{}
	}()

	go func() {
		c, err := discovery.ListenForDiscover()
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println(c)
		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
	case err := <-errChan:
		t.Error(err)
	}
}
