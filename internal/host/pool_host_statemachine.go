package host

import (
	"github.com/filanov/stateswitch"
	"github.com/openshift/assisted-service/models"
)

func NewPoolHostStateMachine(sm stateswitch.StateMachine, th *transitionHandler) stateswitch.StateMachine {

	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeRegisterHost,
		SourceStates: []stateswitch.State{
			"",
		},
		Condition:        th.IsUnboundHost,
		DestinationState: stateswitch.State(models.HostStatusDiscoveringDashUnbound),
		PostTransition:   th.PostRegisterHost,
	})

	// Register host
	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeRegisterHost,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDiscoveringDashUnbound),
			stateswitch.State(models.HostStatusDisconnectedDashUnbound),
			stateswitch.State(models.HostStatusInsufficientDashUnbound),
			stateswitch.State(models.HostStatusKnownDashUnbound),
			stateswitch.State(models.HostStatusUnbinding),
		},
		DestinationState: stateswitch.State(models.HostStatusDiscoveringDashUnbound),
		PostTransition:   th.PostRegisterHost,
	})

	// Disabled host can register if it was booted, no change in the state.
	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType:   TransitionTypeRegisterHost,
		SourceStates:     []stateswitch.State{stateswitch.State(models.HostStatusDisabledDashUnbound)},
		DestinationState: stateswitch.State(models.HostStatusDisabledDashUnbound),
	})

	// Disable host
	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeDisableHost,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDisconnectedDashUnbound),
			stateswitch.State(models.HostStatusDiscoveringDashUnbound),
			stateswitch.State(models.HostStatusInsufficientDashUnbound),
			stateswitch.State(models.HostStatusKnownDashUnbound),
		},
		DestinationState: stateswitch.State(models.HostStatusDisabledDashUnbound),
		PostTransition:   th.PostDisableHost,
	})

	// Enable host
	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeEnableHost,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDisabledDashUnbound),
		},
		DestinationState: stateswitch.State(models.HostStatusDiscoveringDashUnbound),
		PostTransition:   th.PostEnableHost,
	})

	// Bind host
	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeBindHost,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusKnownDashUnbound),
		},
		DestinationState: stateswitch.State(models.HostStatusBinding),
		PostTransition:   th.PostBindHost,
	})

	// Refresh host

	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeRefresh,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDiscoveringDashUnbound),
			stateswitch.State(models.HostStatusInsufficientDashUnbound),
			stateswitch.State(models.HostStatusKnownDashUnbound),
			stateswitch.State(models.HostStatusDisconnectedDashUnbound),
			stateswitch.State(models.HostStatusUnbinding),
		},
		Condition:        stateswitch.Not(If(IsConnected)),
		DestinationState: stateswitch.State(models.HostStatusDisconnectedDashUnbound),
		PostTransition:   th.PostRefreshHost(statusInfoDisconnected),
	})

	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeRefresh,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDisconnectedDashUnbound),
			stateswitch.State(models.HostStatusDiscoveringDashUnbound),
		},
		Condition:        stateswitch.And(If(IsConnected), stateswitch.Not(If(HasInventory))),
		DestinationState: stateswitch.State(models.HostStatusDiscoveringDashUnbound),
		PostTransition:   th.PostRefreshHost(statusInfoDiscovering),
	})

	var hasMinRequiredHardware = stateswitch.And(If(HasMinValidDisks), If(HasMinCPUCores), If(HasMinMemory))
	sufficientToBeBound := stateswitch.And(hasMinRequiredHardware, If(IsNTPSynced))

	// In order for this transition to be fired at least one of the validations in minRequiredHardwareValidations must fail.
	// This transition handles the case that a host does not pass minimum hardware requirements for any of the roles
	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeRefresh,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDisconnectedDashUnbound),
			stateswitch.State(models.HostStatusDiscoveringDashUnbound),
			stateswitch.State(models.HostStatusInsufficientDashUnbound),
			stateswitch.State(models.HostStatusKnownDashUnbound),
		},
		Condition: stateswitch.And(If(IsConnected), If(HasInventory),
			stateswitch.Not(sufficientToBeBound)),
		DestinationState: stateswitch.State(models.HostStatusInsufficientDashUnbound),
		PostTransition:   th.PostRefreshHost(statusInfoInsufficientHardware),
	})

	// Noop transitions
	for _, state := range []stateswitch.State{
		stateswitch.State(models.HostStatusDisabledDashUnbound),
		stateswitch.State(models.HostStatusBinding),
		stateswitch.State(models.HostStatusUnbinding),
	} {
		sm.AddTransition(stateswitch.TransitionRule{
			TransitionType:   TransitionTypeRefresh,
			SourceStates:     []stateswitch.State{state},
			DestinationState: state,
		})
	}

	sm.AddTransition(stateswitch.TransitionRule{
		TransitionType: TransitionTypeRefresh,
		SourceStates: []stateswitch.State{
			stateswitch.State(models.HostStatusDisconnectedDashUnbound),
			stateswitch.State(models.HostStatusDiscoveringDashUnbound),
			stateswitch.State(models.HostStatusInsufficientDashUnbound),
			stateswitch.State(models.HostStatusKnownDashUnbound),
		},
		Condition: stateswitch.And(If(IsConnected), If(HasInventory),
			sufficientToBeBound),
		DestinationState: stateswitch.State(models.HostStatusKnownDashUnbound),
		PostTransition:   th.PostRefreshHost(statusInfoHostReadyToBeBound),
	})

	return sm
}
