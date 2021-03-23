// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/genesis/model"
)

// initSubnet registers subnet endpoints on the given router.
func initSubnet(apiRouter *mux.Router, context *Context) {
	addContext := func(handler contextHandlerFunc) *contextHandler {
		return newContextHandler(context, handler)
	}

	subnetsRouter := apiRouter.PathPrefix("/subnets").Subrouter()
	subnetsRouter.Handle("", addContext(handleGetSubnets)).Methods("GET")

	subnetRouter := apiRouter.PathPrefix("/subnet/{subnet:[A-Za-z0-9]{26}}").Subrouter()
	subnetRouter.Handle("", addContext(handleGetSubnet)).Methods("GET")
}

// handleGetSubnet responds to GET /api/subnet/{subnet}, returning the parent subnet in question.
func handleGetSubnet(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subnetID := vars["subnet"]
	c.Logger = c.Logger.WithField("subnet", subnetID)

	sub, err := c.Store.GetSubnet(subnetID)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query subnet")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if sub == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	outputJSON(c, w, sub)
}

// handleGetSubnets responds to GET /api/subnets, returning the specified page of subnets.
func handleGetSubnets(c *Context, w http.ResponseWriter, r *http.Request) {
	page, perPage, _, freeSubnets, err := parsePaging(r.URL)
	if err != nil {
		c.Logger.WithError(err).Error("failed to parse paging parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := &model.SubnetFilter{
		Page:    page,
		PerPage: perPage,
		Free:    freeSubnets,
	}

	subnets, err := c.Store.GetSubnets(filter)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query subnets")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if subnets == nil {
		subnets = []*model.Subnet{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	outputJSON(c, w, subnets)
}
