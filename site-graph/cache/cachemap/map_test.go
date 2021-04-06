package cachemap

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_cachemap(t *testing.T) {
	type keyVal struct {
		key string
		val string
	}

	tests := []struct {
		name      string
		kv        keyVal
		wantValue interface{}
		wantOk    bool
	}{
		{
			name: "testing lock lock",
			kv: keyVal{
				key: "key",
				val: "val",
			},
			wantValue: "val",
			wantOk:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := New()
			go func() {
				fmt.Println("setting")
				c.Add(tt.kv.key, tt.kv.val)
			}()
			time.Sleep(time.Second * 1)
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					fmt.Println("getting", i)
					gotValue, gotOk := c.Get(tt.kv.key)
					assert.True(t, gotOk)
					assert.Equal(t, tt.wantValue, gotValue)
					fmt.Println("got", i)
				}(i)

			}
			wg.Wait()

		})
	}
}
