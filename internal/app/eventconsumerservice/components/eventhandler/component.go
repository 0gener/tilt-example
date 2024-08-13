package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0gener/go-service/components"
	"github.com/0gener/go-service/lib/awsmessaging"
	"github.com/0gener/tilt-example/internal/app/common"
)

const (
	ComponentName = "eventhandler"
)

type Component struct {
	components.BaseComponent

	messaging *awsmessaging.Component
	queueUrl  string

	sub *awsmessaging.Subscription
}

func New() *Component {
	return &Component{
		BaseComponent: *components.NewBaseComponent(ComponentName),
	}
}

func (component *Component) Configure(_ context.Context) error {
	var err error

	component.messaging, err = components.AsComponent[*awsmessaging.Component](component.Dependency(awsmessaging.ComponentName))
	if err != nil {
		return err
	}

	component.sub = component.messaging.Subscribe(
		component.queueUrl,
		component.handleBatch,
	)

	component.NotifyStatus(components.CONFIGURED)
	return nil
}

func (component *Component) Start(_ context.Context) error {
	err := component.sub.Start()
	if err != nil {
		return err
	}

	component.NotifyStatus(components.STARTED)
	return nil
}

func (component *Component) Shutdown(_ context.Context) error {
	if component.sub != nil {
		component.sub.Stop()
	}

	component.NotifyStatus(components.STOPPED)
	return nil
}

func (component *Component) handleBatch(messages []*awsmessaging.Message) {
	for _, message := range messages {
		message.Err = component.handleSingleMessage(message)
	}
}

func (component *Component) handleSingleMessage(message *awsmessaging.Message) error {
	var event common.ItemCreatedEvent
	if err := json.Unmarshal(message.Data, &event); err != nil {
		return err
	}

	component.Logger().Info(fmt.Sprintf("processed event with ID: %s", event.ID))

	return nil
}
