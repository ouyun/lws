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
	SubMsg           *dbp.Sub
	Subed            bool
}

func newSubscription(subMsg *dbp.Sub) *Subscription {
	closeChan := make(chan struct{})
	notificationChan := make(chan *Notification)
	return &Subscription{
		CloseChan:        closeChan,
		NotificationChan: notificationChan,
		SubMsg:           subMsg,
	}
}

// subscribe topic
// don't forget to call deleteSubscription if sub failed or unsubcribe
func (c *Client) Subscribe(subMsg *dbp.Sub) (*Subscription, interface{}, error) {
	var err error
	var response interface{}
	// 1. generate unique id, considering the re-dial logic
	subMsg.Id = generateUuidString()
	// subMsg.Id = "3494"

	// 2. make Subscription and add it to map
	subscription := newSubscription(subMsg)

	c.subLock.Lock()
	if _, ok := c.subscriptions[subMsg.Id]; ok {
		// TODO should we delete old one and re-sub it? Consider the stacked Notification in old chan
		err = fmt.Errorf("Sub Id [%s] is already existed", subMsg.Id)
		return nil, nil, err
	}

	c.subscriptions[subMsg.Id] = subscription
	c.subLock.Unlock()

	// 3. send sub request to core wallet
	response, err = c.Call(subMsg)
	for ; IsClientTimeoutError(err); response, err = c.Call(subMsg) {
		c.LogError("subscribe [%s] timeout, retry", subMsg.Name)
	}

	if err != nil {
		c.deleteSubscription(subMsg.Id)
	}

	_, ok := response.(*dbp.Nosub)
	if ok {
		c.deleteSubscription(subMsg.Id)
	} else if err == nil {
		// success
		c.LogError("sub [%s][%s] done", subMsg.Name, subMsg.Id)
		c.markSubDone(subMsg.Id)
	}
	return subscription, response, err
}

func (c *Client) Resubscribe() {
	var response interface{}
	var err error
	c.subLock.Lock()
	for _, sub := range c.subscriptions {
		// only resub subscribed msgs, ignore the hanging ones
		if sub.Subed {
			subMsg := sub.SubMsg
			c.LogError("Resubscribe Name[%s] Id[%s]", subMsg.Name, subMsg.Id)

			// do not retry
			response, err = c.Call(subMsg)

			if err != nil {
				c.LogError("re-subscribe [%s] [%s] timeout, retry", subMsg.Name, subMsg.Id)
				c.deleteSubscription(subMsg.Id)
			}

			_, ok := response.(*dbp.Nosub)
			if ok {
				c.deleteSubscription(subMsg.Id)
			}
		}
	}
	c.subLock.Unlock()
}

func (c *Client) markSubDone(subId string) error {
	c.subLock.Lock()
	subscription, ok := c.subscriptions[subId]
	if !ok {
		return nil
	}
	subscription.Subed = true

	c.subLock.Unlock()
	return nil
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
