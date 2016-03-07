package voyage

import "github.com/go-kit/kit/examples/shipping/location"

// A set of sample voyages.
var (
	V100 = New("V100", Schedule{
		[]CarrierMovement{
			{DepartureLocation: location.Hongkong, ArrivalLocation: location.Tokyo},
			{DepartureLocation: location.Tokyo, ArrivalLocation: location.NewYork},
		},
	})

	V300 = New("V300", Schedule{
		[]CarrierMovement{
			{DepartureLocation: location.Tokyo, ArrivalLocation: location.Rotterdam},
			{DepartureLocation: location.Rotterdam, ArrivalLocation: location.Hamburg},
			{DepartureLocation: location.Hamburg, ArrivalLocation: location.Melbourne},
			{DepartureLocation: location.Melbourne, ArrivalLocation: location.Tokyo},
		},
	})

	V400 = New("V400", Schedule{
		[]CarrierMovement{
			{DepartureLocation: location.Hamburg, ArrivalLocation: location.Stockholm},
			{DepartureLocation: location.Stockholm, ArrivalLocation: location.Helsinki},
			{DepartureLocation: location.Helsinki, ArrivalLocation: location.Hamburg},
		},
	})
)

// These voyages are hard-coded into the current pathfinder. Make sure
// they exist.
var (
	V0100S = New("0100S", Schedule{[]CarrierMovement{}})
	V0200T = New("0200T", Schedule{[]CarrierMovement{}})
	V0300A = New("0300A", Schedule{[]CarrierMovement{}})
	V0301S = New("0301S", Schedule{[]CarrierMovement{}})
	V0400S = New("0400S", Schedule{[]CarrierMovement{}})
)
