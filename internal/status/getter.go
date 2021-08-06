package status

import (
	"context"
	"github.com/armory/armory-cli/internal/deng"
)

type eventGetter func(ctx context.Context) ([]*deng.Event, error)

func noopEventGetter(ctx context.Context) ([]*deng.Event, error) {
	return nil, nil
}

type onceEventGetter struct {
	client       deng.DeploymentServiceClient
	deploymentId string
}

func (o *onceEventGetter) getEvents(ctx context.Context) ([]*deng.Event, error) {
	eResp, err := o.client.GetEvents(ctx, &deng.GetEventRequest{
		DeploymentId: o.deploymentId,
	})
	if err != nil || eResp == nil {
		return nil, err
	}
	return eResp.Events, nil
}

type watchGetter struct {
	client     deng.DeploymentServiceClient
	descriptor *deng.Descriptor
	events     []*deng.Event
	lastId     int64
}

func (w *watchGetter) tick(ctx context.Context) error {
	res, err := w.client.GetEvents(ctx, &deng.GetEventRequest{
		DeploymentId: w.descriptor.Id,
		EventFilter: &deng.GetEventRequest_EventFilter{
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

func (w *watchGetter) getEvents(ctx context.Context) ([]*deng.Event, error) {
	return w.events, nil
}
