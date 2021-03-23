// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

const (
	// AccountStateStable is a account in a stable state and undergoing no changes.
	AccountStateStable = "stable"
	// AccountStateCreationRequested is a account in the process of being created.
	AccountStateCreationRequested = "creation-requested"
	// AccountStateCreationFailed is a account that failed creation.
	AccountStateCreationFailed = "creation-failed"
	// AccountStateProvisioningRequested is a account in the process of being provisioned.
	AccountStateProvisioningRequested = "provisioning-requested"
	// AccountStateRefreshMetadata is a account that will have metadata refreshed.
	AccountStateRefreshMetadata = "refresh-metadata"
	// AccountStateProvisioningFailed is a account that failed provisioning.
	AccountStateProvisioningFailed = "provisioning-failed"
	// AccountStateDeletionRequested is a account in the process of being deleted.
	AccountStateDeletionRequested = "deletion-requested"
	// AccountStateDeletionFailed is a account that failed deletion.
	AccountStateDeletionFailed = "deletion-failed"
	// AccountStateDeleted is a account that has been deleted
	AccountStateDeleted = "deleted"
)

// AllAccountStates is a list of all states a account can be in.
// Warning:
// When creating a new account state, it must be added to this list.
var AllAccountStates = []string{
	AccountStateStable,
	AccountStateRefreshMetadata,
	AccountStateCreationRequested,
	AccountStateCreationFailed,
	AccountStateProvisioningRequested,
	AccountStateProvisioningFailed,
	AccountStateDeletionRequested,
	AccountStateDeletionFailed,
	AccountStateDeleted,
}

// AllAccountStatesPendingWork is a list of all account states that the supervisor
// will attempt to transition towards stable on the next "tick".
// Warning:
// When creating a new account state, it must be added to this list if the cloud
// account supervisor should perform some action on its next work cycle.
var AllAccountStatesPendingWork = []string{
	AccountStateCreationRequested,
	AccountStateProvisioningRequested,
	AccountStateRefreshMetadata,
	AccountStateDeletionRequested,
}

// AllAccountRequestStates is a list of all states that a account can be put in
// via the API.
// Warning:
// When creating a new account state, it must be added to this list if an API
// endpoint should put the account in this state.
var AllAccountRequestStates = []string{
	AccountStateCreationRequested,
	AccountStateProvisioningRequested,
	AccountStateDeletionRequested,
}

// ValidTransitionState returns whether a account can be transitioned into the
// new state or not based on its current state.
func (c *Account) ValidTransitionState(newState string) bool {
	switch newState {
	case AccountStateCreationRequested:
		return validTransitionToAccountStateCreationRequested(c.State)
	case AccountStateProvisioningRequested:
		return validTransitionToAccountStateProvisioningRequested(c.State)
	case AccountStateDeletionRequested:
		return validTransitionToAccountStateDeletionRequested(c.State)
	}

	return false
}

func validTransitionToAccountStateCreationRequested(currentState string) bool {
	switch currentState {
	case AccountStateCreationRequested,
		AccountStateCreationFailed:
		return true
	}

	return false
}

func validTransitionToAccountStateProvisioningRequested(currentState string) bool {
	switch currentState {
	case AccountStateStable,
		AccountStateProvisioningFailed,
		AccountStateProvisioningRequested:
		return true
	}

	return false
}

func validTransitionToAccountStateDeletionRequested(currentState string) bool {
	switch currentState {
	case AccountStateStable,
		AccountStateCreationRequested,
		AccountStateCreationFailed,
		AccountStateProvisioningFailed,
		AccountStateDeletionRequested,
		AccountStateDeletionFailed:
		return true
	}

	return false
}

// AccountStateReport is a report of all account requests states.
type AccountStateReport []StateReportEntry

// StateReportEntry is a report entry of a given request state.
type StateReportEntry struct {
	RequestedState string
	ValidStates    StateList
	InvalidStates  StateList
}

// StateList is a list of states
type StateList []string

// Count provides the number of states in a StateList.
func (sl *StateList) Count() int {
	return len(*sl)
}

// GetAccountRequestStateReport returns a AccountStateReport based on the current
// model of account states.
func GetAccountRequestStateReport() AccountStateReport {
	report := AccountStateReport{}

	for _, requestState := range AllAccountRequestStates {
		entry := StateReportEntry{
			RequestedState: requestState,
		}

		for _, newState := range AllAccountStates {
			c := Account{State: newState}
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
