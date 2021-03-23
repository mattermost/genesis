// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

const (
	// ParentSubnetStateAdded is a parent subnet in an added state and undergoing no changes.
	ParentSubnetStateAdded = "added"
	// ParentSubnetStateAdditionRequested is a parent subnet in the process of being added.
	ParentSubnetStateAdditionRequested = "addition-requested"
	// ParentSubnetStateAdditionFailed is a parent subnet that failed creation.
	ParentSubnetStateAdditionFailed = "addition-failed"
	// ParentSubnetStateSplitRequested is a parent subnet in the process of being split into subnets.
	ParentSubnetStateSplitRequested = "split-requested"
	// ParentSubnetStateSplitFailed is a parent subnet that failed being split.
	ParentSubnetStateSplitFailed = "split-failed"
)

// AllParentSubnetStates is a list of all states a parent subnet can be in.
// Warning:
// When creating a new parent subnet state, it must be added to this list.
var AllParentSubnetStates = []string{
	ParentSubnetStateAdded,
	ParentSubnetStateAdditionRequested,
	ParentSubnetStateAdditionFailed,
	ParentSubnetStateSplitRequested,
	ParentSubnetStateSplitFailed,
}

// AllParentSubnetStatesPendingWork is a list of all parent subnet states that the supervisor
// will attempt to transition towards added on the next "tick".
// Warning:
// When creating a new parent subnet state, it must be added to this list if the cloud
// parent subnet supervisor should perform some action on its next work cycle.
var AllParentSubnetStatesPendingWork = []string{
	ParentSubnetStateAdditionRequested,
	ParentSubnetStateSplitRequested,
}

// AllParentSubnetRequestStates is a list of all states that a parent subnet can be put in
// via the API.
// Warning:
// When creating a new parent subnet state, it must be added to this list if an API
// endpoint should put the parent subnet in this state.
var AllParentSubnetRequestStates = []string{
	ParentSubnetStateAdditionRequested,
}

// ValidTransitionState returns whether a parent subnet can be transitioned into the
// new state or not based on its current state.
func (c *ParentSubnet) ValidTransitionState(newState string) bool {
	switch newState {
	case ParentSubnetStateAdditionRequested:
		return validTransitionToParentSubnetStateAdditionRequested(c.State)
	case ParentSubnetStateSplitRequested:
		return validTransitionToParentSubnetStateSplitRequested(c.State)
	}

	return false
}

func validTransitionToParentSubnetStateAdditionRequested(currentState string) bool {
	switch currentState {
	case ParentSubnetStateAdditionRequested,
		ParentSubnetStateAdditionFailed:
		return true
	}

	return false
}

func validTransitionToParentSubnetStateSplitRequested(currentState string) bool {
	switch currentState {
	case ParentSubnetStateAdded,
		ParentSubnetStateSplitFailed,
		ParentSubnetStateSplitRequested:
		return true
	}

	return false
}

// ParentSubnetStateReport is a report of all account requests states.
type ParentSubnetStateReport []StateReportEntry

// GetParentSubnetRequestStateReport returns a AccountStateReport based on the current
// model of account states.
func GetParentSubnetRequestStateReport() ParentSubnetStateReport {
	report := ParentSubnetStateReport{}

	for _, requestState := range AllParentSubnetRequestStates {
		entry := StateReportEntry{
			RequestedState: requestState,
		}

		for _, newState := range AllParentSubnetStates {
			c := ParentSubnet{State: newState}
			if c.ValidTransitionState(requestState) {
				entry.ValidStates = append(entry.ValidStates, newState)
			} else {
				entry.InvalidStates = append(entry.InvalidStates, newState)
			}
		}

		report = append(report, entry)
	}

	return report
}
