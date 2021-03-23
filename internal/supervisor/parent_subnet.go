// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package supervisor

import (
	"fmt"
	"net"
	"time"

	"github.com/mattermost/genesis/internal/webhook"

	"github.com/mattermost/genesis/model"
	log "github.com/sirupsen/logrus"
)

// parentSubnetStore abstracts the database operations required to query parent subnets.
type parentSubnetStore interface {
	GetParentSubnet(subnet string) (*model.ParentSubnet, error)
	GetUnlockedParentSubnetsPendingWork() ([]*model.ParentSubnet, error)
	GetParentSubnets(ParentSubnetFilter *model.ParentSubnetFilter) ([]*model.ParentSubnet, error)
	UpdateParentSubnet(parentSubnet *model.ParentSubnet) error
	LockParentSubnet(subnet, lockerID string) (bool, error)
	UnlockParentSubnet(subnet string, lockerID string, force bool) (bool, error)

	GetWebhooks(filter *model.WebhookFilter) ([]*model.Webhook, error)

	AddSubnet(subnet *model.Subnet) error
}

// parentSubnetProvisioner abstracts the provisioning operations required by the parent subnet supervisor.
type parentSubnetProvisioner interface {
	AddParentSubnet(parentSubnet *model.ParentSubnet) error
	SplitParentSubnet(parentSubnet *model.ParentSubnet) ([]net.IPNet, error)
}

// ParentSubnetSupervisor finds parent subnets pending work and effects the required changes.
//
// The degree of parallelism is controlled by a weighted semaphore, intended to be shared with
// other clients needing to coordinate background jobs.
type ParentSubnetSupervisor struct {
	store       parentSubnetStore
	provisioner parentSubnetProvisioner
	instanceID  string
	logger      log.FieldLogger
}

// NewParentSubnetSupervisor creates a new ParentSubnetSupervisor.
func NewParentSubnetSupervisor(store parentSubnetStore, parentSubnetProvisioner parentSubnetProvisioner, instanceID string, logger log.FieldLogger) *ParentSubnetSupervisor {
	return &ParentSubnetSupervisor{
		store:       store,
		provisioner: parentSubnetProvisioner,
		instanceID:  instanceID,
		logger:      logger,
	}
}

// Shutdown performs graceful shutdown tasks for the parent subnet supervisor.
func (s *ParentSubnetSupervisor) Shutdown() {
	s.logger.Debug("Shutting down parent subnet supervisor")
}

// Do looks for work to be done on any pending parent subnets and attempts to schedule the required work.
func (s *ParentSubnetSupervisor) Do() error {
	parentSubnets, err := s.store.GetUnlockedParentSubnetsPendingWork()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to query for parent subnets pending work")
		return nil
	}

	for _, parentSubnet := range parentSubnets {
		s.Supervise(parentSubnet)
	}

	return nil
}

// Supervise schedules the required work on the given parent subnet.
func (s *ParentSubnetSupervisor) Supervise(parentSubnet *model.ParentSubnet) {
	logger := s.logger.WithFields(log.Fields{
		"subnet": parentSubnet.ID,
	})

	lock := newParentSubnetLock(parentSubnet.ID, s.instanceID, s.store, logger)
	if !lock.TryLock() {
		return
	}
	defer lock.Unlock()

	// Before working on the parent subnet, it is crucial that we ensure that it was
	// not updated to a new state by another genesis server.
	originalState := parentSubnet.State
	parentSubnet, err := s.store.GetParentSubnet(parentSubnet.ID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get refreshed parent subnet")
		return
	}
	if parentSubnet.State != originalState {
		logger.WithField("oldParentSubnetState", originalState).
			WithField("newParentSubnetState", parentSubnet.State).
			Warn("Another provisioner has worked on this parent subnet; skipping...")
		return
	}

	logger.Debugf("Supervising parent subnet in state %s", parentSubnet.State)

	newState := s.transitionParentSubnet(parentSubnet, logger)

	parentSubnet, err = s.store.GetParentSubnet(parentSubnet.ID)
	if err != nil {
		logger.WithError(err).Warnf("failed to get parent subnet and thus persist state %s", newState)
		return
	}

	if parentSubnet.State == newState {
		return
	}

	oldState := parentSubnet.State
	parentSubnet.State = newState
	err = s.store.UpdateParentSubnet(parentSubnet)
	if err != nil {
		logger.WithError(err).Warnf("failed to set parent subnet state to %s", newState)
		return
	}

	webhookPayload := &model.WebhookPayload{
		Type:      model.TypeParentSubnet,
		ID:        parentSubnet.ID,
		NewState:  newState,
		OldState:  oldState,
		Timestamp: time.Now().UnixNano(),
		ExtraData: map[string]string{},
	}
	err = webhook.SendToAllWebhooks(s.store, webhookPayload, logger.WithField("webhookEvent", webhookPayload.NewState))
	if err != nil {
		logger.WithError(err).Error("Unable to process and send webhooks")
	}

	logger.Debugf("Transitioned parent subnet from %s to %s", oldState, newState)
}

// Do works with the given parent subnet to transition it to a final state.
func (s *ParentSubnetSupervisor) transitionParentSubnet(parentSubnet *model.ParentSubnet, logger log.FieldLogger) string {
	switch parentSubnet.State {
	case model.ParentSubnetStateAdditionRequested:
		return s.addParentSubnet(parentSubnet, logger)
	default:
		logger.Warnf("Found parent subnet pending work in unexpected state %s", parentSubnet.State)
		return parentSubnet.State
	}
}

func (s *ParentSubnetSupervisor) addParentSubnet(parentSubnet *model.ParentSubnet, logger log.FieldLogger) string {
	var err error

	err = s.provisioner.AddParentSubnet(parentSubnet)
	if err != nil {
		logger.WithError(err).Error("Failed to add parent subnet")
		return model.ParentSubnetStateAdditionFailed
	}

	logger.Info("Finished adding parent subnet")
	return s.splitParentSubnets(parentSubnet, logger)
}

func (s *ParentSubnetSupervisor) splitParentSubnets(parentSubnet *model.ParentSubnet, logger log.FieldLogger) string {
	subnets, err := s.provisioner.SplitParentSubnet(parentSubnet)
	if err != nil {
		logger.WithError(err).Error("Failed to split parent subnet")
		return model.ParentSubnetStateSplitFailed
	}

	for _, subnet := range subnets {
		sub := model.Subnet{
			CIDR:           fmt.Sprintf("%s/%d", &subnet.IP, parentSubnet.SplitRange),
			Used:           false,
			ParentSubnet:   parentSubnet.CIDR,
			SubnetMetadata: &model.SubnetMetadata{},
			CreateAt:       parentSubnet.CreateAt,
		}
		err = s.store.AddSubnet(&sub)
		if err != nil {
			logger.WithError(err).Error("failed to add subnet")
			return model.ParentSubnetStateSplitFailed
		}
	}

	logger.Info("Finished splitting parent subnet")
	return model.ParentSubnetStateAdded
}
