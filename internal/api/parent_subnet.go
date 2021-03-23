// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/genesis/internal/webhook"
	"github.com/mattermost/genesis/model"
)

// initParentSubnet registers parent subnet endpoints on the given router.
func initParentSubnet(apiRouter *mux.Router, context *Context) {
	addContext := func(handler contextHandlerFunc) *contextHandler {
		return newContextHandler(context, handler)
	}

	parentSubnetsRouter := apiRouter.PathPrefix("/parentsubnets").Subrouter()
	parentSubnetsRouter.Handle("", addContext(handleGetParentSubnets)).Methods("GET")
	parentSubnetsRouter.Handle("", addContext(handleAddParentSubnet)).Methods("POST")

	parentSubnetRouter := apiRouter.PathPrefix("/parentsubnet/{parentsubnet:[A-Za-z0-9]{26}}").Subrouter()
	parentSubnetRouter.Handle("", addContext(handleGetParentSubnet)).Methods("GET")
}

// handleGetParentSubnet responds to GET /api/parentsubnet/{parentsubnet}, returning the parent subnet in question.
func handleGetParentSubnet(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subnetID := vars["parentsubnet"]
	c.Logger = c.Logger.WithField("parent-subnet", subnetID)

	parentSubnet, err := c.Store.GetParentSubnet(subnetID)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query parent subnet")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if parentSubnet == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	outputJSON(c, w, parentSubnet)
}

// handleGetParentSubnets responds to GET /api/parentsubnets, returning the specified page of parent subnets.
func handleGetParentSubnets(c *Context, w http.ResponseWriter, r *http.Request) {
	page, perPage, _, _, err := parsePaging(r.URL)
	if err != nil {
		c.Logger.WithError(err).Error("failed to parse paging parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := &model.ParentSubnetFilter{
		Page:    page,
		PerPage: perPage,
	}

	parentSubnets, err := c.Store.GetParentSubnets(filter)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query parent subnets")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if parentSubnets == nil {
		parentSubnets = []*model.ParentSubnet{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	outputJSON(c, w, parentSubnets)
}

// handleAddParentSubnet responds to POST /api/parentsubnets, beginning the process of creating a new parent subnet.
func handleAddParentSubnet(c *Context, w http.ResponseWriter, r *http.Request) {
	addParentSubnetRequest, err := model.NewAddParentSubnetRequestFromReader(r.Body)
	if err != nil {
		c.Logger.WithError(err).Error("failed to decode request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parentSubnet := model.ParentSubnet{
		CIDR:       addParentSubnetRequest.CIDR,
		SplitRange: addParentSubnetRequest.SplitRange,
		State:      model.ParentSubnetStateAdditionRequested,
	}

	err = c.Store.AddParentSubnet(&parentSubnet)
	if err != nil {
		c.Logger.WithError(err).Error("failed to create account")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	webhookPayload := &model.WebhookPayload{
		Type:      model.TypeParentSubnet,
		ID:        parentSubnet.ID,
		NewState:  model.ParentSubnetStateAdditionRequested,
		OldState:  "n/a",
		Timestamp: time.Now().UnixNano(),
		ExtraData: map[string]string{"CIDR": parentSubnet.CIDR},
	}
	err = webhook.SendToAllWebhooks(c.Store, webhookPayload, c.Logger.WithField("webhookEvent", webhookPayload.NewState))
	if err != nil {
		c.Logger.WithError(err).Error("Unable to process and send webhooks")
	}

	c.Supervisor.Do()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	outputJSON(c, w, parentSubnet)
}

// handleRetryAddParentSubnet responds to POST /api/parentsubnet/{parentsubnet}, retrying a previously
// failed creation.
func handleRetryAddParentSubnet(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentSubnet := vars["parentsubnet"]
	c.Logger = c.Logger.WithField("parent-subnet", parentSubnet)

	parentSub, status, unlockOnce := lockParentSubnet(c, parentSubnet)
	if status != 0 {
		w.WriteHeader(status)
		return
	}
	defer unlockOnce()

	newState := model.ParentSubnetStateAdditionRequested

	if !parentSub.ValidTransitionState(newState) {
		c.Logger.Warnf("unable to retry parent subnet creation while in state %s", parentSub.State)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if parentSub.State != newState {
		webhookPayload := &model.WebhookPayload{
			Type:      model.TypeAccount,
			ID:        parentSub.ID,
			NewState:  newState,
			OldState:  parentSub.State,
			Timestamp: time.Now().UnixNano(),
			ExtraData: map[string]string{},
		}
		parentSub.State = newState

		err := c.Store.UpdateParentSubnet(parentSub)
		if err != nil {
			c.Logger.WithError(err).Errorf("failed to retry parent subnet creation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = webhook.SendToAllWebhooks(c.Store, webhookPayload, c.Logger.WithField("webhookEvent", webhookPayload.NewState))
		if err != nil {
			c.Logger.WithError(err).Error("Unable to process and send webhooks")
		}
	}

	// Notify even if we didn't make changes, to expedite even the no-op operations above.
	unlockOnce()
	c.Supervisor.Do()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	outputJSON(c, w, parentSub)
}
