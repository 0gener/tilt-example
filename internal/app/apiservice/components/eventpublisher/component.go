package eventpublisher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0gener/go-service/components"
	"github.com/0gener/go-service/lib/awsmessaging"
	"github.com/0gener/tilt-example/internal/app/common"
)

const (
	ComponentName = "eventpublisher"
)

type Component struct {
	components.BaseComponent

	messaging *awsmessaging.Component
	topicArn  string
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

	component.NotifyStatus(components.CONFIGURED)
	return nil
}

func (component *Component) ItemCreatedEvent(ctx context.Context, event common.ItemCreatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ItemCreatedEvent: %w", err)
	}

	msg := awsmessaging.NewMessage(data,
		awsmessaging.WithAttribute("resource_type", "items"),
		awsmessaging.WithAttribute("event_type", "created"),
	)
	return component.messaging.Publish(ctx, component.topicArn, msg)
}
