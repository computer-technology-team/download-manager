package downloads

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
	ticker := time.Tick(time.Duration(t.tickerDelay) * time.Second)
	for {
		select {
		case <-t.generationQuiteChan:
			return
		case <-ticker:
			t.tokens <- 0
		}
	}
}

func (t *Ticker) generate() {
	for {
		select {
		case <-t.generationQuiteChan:
			close(t.tokens)
			return
		default:
			fmt.Println("ticker delay: ", t.tickerDelay)
			t.tokens <- 0
			time.Sleep(time.Duration(t.tickerDelay * 1e9))
		}
	}
}

func (t *Ticker) Start() {
	go t.generate()
}

func (t *Ticker) GetToken() {// TODO don't block if BW is limitless
	<-t.tokens
}
func (t *Ticker) Quite() {//TODO handle limitless
	t.generationQuiteChan <- 0
}
