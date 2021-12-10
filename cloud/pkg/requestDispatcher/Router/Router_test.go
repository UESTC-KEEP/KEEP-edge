package Router

import (
	"keep/pkg/util/core/model"
	"testing"
	"time"
)

func TestRouter_MessageDispatcher(t *testing.T) {

	msg := &model.Message{}
	msg.Content = "hello kafka"
	msg.Router.Group = "/log"

	msg1 := &model.Message{}
	msg1.Content = "hello kafka1"
	msg1.Router.Group = "/add"

	go MessageRouter()

	RevMsgChan <- msg

	time.Sleep(3 * time.Second)
	RevMsgChan <- msg1

}

func TestRouter_MessageDispatcher1(t *testing.T) {

	msg := &model.Message{}
	msg.Content = "hello kafka"
	msg.Router.Group = "/log"

	msg1 := &model.Message{}
	msg1.Content = "hello kafka1"
	msg1.Router.Group = "/log"

	go MessageRouter()

	RevMsgChan <- msg

	time.Sleep(3 * time.Second)
	RevMsgChan <- msg1

}