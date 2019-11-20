package pipeline_test

import (
	"testing"

	"github.com/Rennbon/boxwallet/bccore"
	"github.com/Rennbon/boxwallet/pipeline"
)

var hCache = pipeline.GetHeightCacheInstance()

func TestHeightCache_Push(t *testing.T) {
	hCache.Push(bccore.BC_BTC, 1000, 500)
	hs := hCache.LoadAll()
	for k, v := range hs {
		t.Log(k)
		t.Log(v.PubHeight, v.CurHeight)
	}
}
