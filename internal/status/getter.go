package status

import (
	"context"
	"github.com/armory/armory-cli/internal/deng/protobuff"
)

type eventGetter func(ctx context.Context) ([]*protobuff.Event, error)

func noopEventGetter(ctx context.Context) ([]*protobuff.Event, error) {
	return nil, nil
}

type onceEventGetter struct {
	client       protobuff.DeploymentServiceClient
	deploymentId string
}

func (o *onceEventGetter) getEvents(ctx context.Context) ([]*protobuff.Event, error) {
	eResp, err := o.client.GetEvents(ctx, &protobuff.GetEventRequest{
		DeploymentId: o.deploymentId,
	})
	if err != nil || eResp == nil {
		return nil, err
	}
	return eResp.Events, nil
}

type watchGetter struct {
	client     protobuff.DeploymentServiceClient
	descriptor *protobuff.Descriptor
	events     []*protobuff.Event
	lastId     int64
}

func (w *watchGetter) tick(ctx context.Context) error {
	res, err := w.client.GetEvents(ctx, &protobuff.GetEventRequest{
		DeploymentId: w.descriptor.Id,
		EventFilter: &protobuff.GetEventRequest_EventFilter{
			AfterId: w.lastId,
		},
	})
	if err != nil {
		return err
	}
	for _, e := range res.Events {
		w.events = append(w.events, e)
		if e.Status != w.descriptor.Status {
			w.descriptor.Status = e.Status
		}
	}
	w.lastId = res.Cursor
	return nil
}

func (w *watchGetter) getEvents(ctx context.Context) ([]*protobuff.Event, error) {
	return w.events, nil
}
