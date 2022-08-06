package types

import "fmt"

func NewGenesisState(params Params, zoneDrops []*ZoneDrop, claimRecords []*ClaimRecord) *GenesisState {
	return &GenesisState{
		Params:       params,
		ZoneDrops:    zoneDrops,
		ClaimRecords: claimRecords,
	}
}

// DefaultGenesis returns the default ics genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		ZoneDrops:    make([]*ZoneDrop, 0),
		ClaimRecords: make([]*ClaimRecord, 0),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants hold.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	zoneMap := make(map[string]int)
	sumMap := make(map[string]uint64)
	for i, zd := range data.ZoneDrops {
		// check for duplicate zone chain id
		if zdi, exists := zoneMap[zd.ChainId]; exists {
			return fmt.Errorf("%w, [%d] %s already used for zone drop [%d]", ErrDuplicateZoneDrop, i, zd.ChainId, zdi)
		}
		// validate zone drop
		if err := zd.ValidateBasic(); err != nil {
			return err
		}
		// add to lookup map
		zoneMap[zd.ChainId] = i
		sumMap[zd.ChainId] = 0
	}

	claimMap := make(map[string]int)
	for i, cr := range data.ClaimRecords {
		// validate claim record
		if err := cr.ValidateBasic(); err != nil {
			return err
		}
		// check for duplicate
		key := cr.ChainId + "." + cr.Address
		if cmi, exists := claimMap[key]; exists {
			return fmt.Errorf("%w, [%d] %s already used for zone drop [%d]", ErrDuplicateClaimRecord, i, key, cmi)
		}
		claimMap[key] = i
		// sum MaxAllocations per zone
		if _, exists := zoneMap[cr.ChainId]; !exists {
			return fmt.Errorf("%w, %s for claim record [%d]", ErrZoneDropNotFound, cr.ChainId, i)
		}
		sumMap[cr.ChainId] += cr.MaxAllocation
	}

	for _, zd := range data.ZoneDrops {
		if zd.Allocation < sumMap[zd.ChainId] {
			return ErrAllocationExceeded
		}
	}

	return nil
}
