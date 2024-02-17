package main

import (
	"concurrency/producer_consumer/message"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	mr := message.NewMessenger(5, &wg)

	<-mr.Produce(10)
	<-mr.Consume()

	wg.Wait()
}
