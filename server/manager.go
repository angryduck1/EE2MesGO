package server

const FixedQueueLimit = 500

var ResponseQueue chan interface{}
var semaphore = make(chan struct{}, FixedQueueLimit)

func manageActivity() {
	for msg := range ResponseQueue {
		switch msg.(type) {

		case SyncInfo:
			semaphore <- struct{}{}
		}

		defer func() {
			<-semaphore
		}()
	}
}
