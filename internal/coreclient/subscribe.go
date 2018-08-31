package coreclient

import (
	"fmt"
	// "time"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
)

type Notification struct {
	Msg interface{}
	Id  string
	// ReceivedAt time.Time
}

type Subscription struct {
	CloseChan        chan struct{}
	NotificationChan chan *Notification
}

func newSubscription() *Subscription {
	closeChan := make(chan struct{})
	notificationChan := make(chan *Notification)
	return &Subscription{
		CloseChan:        closeChan,
		NotificationChan: notificationChan,
	}
}

// subscribe topic
// don't forget to call deleteSubscription if sub failed or unsubcribe
func (c *Client) Subscribe(subMsg *dbp.Sub) (*Subscription, interface{}, error) {
	// 1. generate unique id, considering the re-dial logic
	subMsg.Id = generateUuidString()

	// 2. make Subscription and add it to map
	subscription := newSubscription()

	// wait until c.subscriptions map is available
	// for {
	// 	if c.subscriptions != nil {
	// 		break
	// 	}
	// 	select {
	// 	case <-time.After(time.Second):
	// 		continue
	// 	}
	// }

	c.subLock.Lock()
	if _, ok := c.subscriptions[subMsg.Id]; ok {
		// TODO should we delete old one and re-sub it? Consider the stacked Notification in old chan
		err := fmt.Errorf("Sub Id [%s] is already existed", subMsg.Id)
		return nil, nil, err
	}

	c.subscriptions[subMsg.Id] = subscription
	c.subLock.Unlock()

	// 3. send sub request to core wallet
	response, err := c.Call(subMsg)
	if err != nil {
		c.deleteSubscription(subMsg.Id)
	}
	return subscription, response, err
}

func (c *Client) deleteSubscription(subId string) error {
	// 1. delete subscription from map
	c.subLock.Lock()
	subscription, ok := c.subscriptions[subId]
	if !ok {
		return nil
	}

	delete(c.subscriptions, subId)
	c.subLock.Unlock()

	// 2. close chan
	close(subscription.CloseChan)

	return nil
}

func (c *Client) handleNotification(subId string, response interface{}) {
	c.LogDebug("recevied notification id [%s]", subId)
	// 1. find Subscription from map by id
	subscription, ok := c.subscriptions[subId]
	if !ok {
		c.LogError("unknown notificaiotn id: [%s] received", subId)
		return
	}

	// 2. pack Notification
	notification := &Notification{
		Msg: response,
		Id:  subId,
	}

	// 3. send Notification to chan
	subscription.NotificationChan <- notification
}
