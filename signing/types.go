package signing

import (
	boostTypes "github.com/flashbots/go-boost-utils/types"
)

type (
	Domain      [32]byte
	DomainType  [4]byte
	ForkVersion [4]byte
	Root        [32]byte
)

func (d *Domain) ToFlashbots() boostTypes.Domain {
	boostDomain := new(boostTypes.Domain)
	copy(boostDomain[:], d[:])
	return *boostDomain
}
