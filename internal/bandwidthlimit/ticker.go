package bandwidthlimit

import (
	"fmt"
	"sync"
	"time"
)

const MaxpacketSize int64 = 65535
const defaultBandWidth int64 = 1e6

type Ticker struct {
	mutex               *sync.Mutex
	tickerDelay         float64
	tokens              chan interface{}
	generationQuiteChan chan interface{}
}

func NewTicker() Ticker {
	t := Ticker{
		mutex:               &sync.Mutex{},
		tickerDelay:         0.,
		tokens:              make(chan interface{}, 100),
		generationQuiteChan: make(chan interface{}, 1),
	}
	t.SetBandwidth(defaultBandWidth)
	return t
}
func (t *Ticker) SetBandwidth(BandwidthLimitBytesPS int64) {
	t.mutex.Lock()
	t.tickerDelay = float64(MaxpacketSize) / float64(BandwidthLimitBytesPS)
	t.mutex.Unlock()
}
func (t *Ticker) generate() {
	for {
		select {
		case <-t.generationQuiteChan:
			fmt.Println("stoppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp")
			return
		default:
			t.tokens <- 0
			time.Sleep(time.Second * time.Duration(t.tickerDelay))
		}
	}
}

func (t *Ticker) Start() {
	go t.generate()
}

func (t *Ticker) GetToken() { // TODO don't block if BW is limitless
	<-t.tokens
}
func (t *Ticker) Stop() { //TODO handle limitless
	t.generationQuiteChan <- 0
}
