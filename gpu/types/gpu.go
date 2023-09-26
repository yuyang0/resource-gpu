package types

import (
	"strings"

	"github.com/cockroachdb/errors"
)

type ProdCountMap map[string]int

func (pcm ProdCountMap) Validate() error {
	for prod, count := range pcm {
		if count <= 0 {
			return errors.Wrapf(ErrInvalidGPUMap, "count is less or equal to zero")
		}
		if strings.Trim(prod, " ") == "" {
			return errors.Wrapf(ErrInvalidGPUProduct, "product is empty")
		}
	}
	return nil
}

func (pcm ProdCountMap) ValidateProd() error {
	for prod := range pcm {
		if strings.Trim(prod, " ") == "" {
			return errors.Wrapf(ErrInvalidGPUProduct, "product is empty")
		}
	}
	return nil
}

func (pcm ProdCountMap) ValidateCount() error {
	for prod, count := range pcm {
		if count <= 0 {
			return errors.Wrapf(ErrInvalidGPUMap, "%s: count is less or equal to zero", prod)
		}
	}
	return nil
}

func (pcm ProdCountMap) Add(g1 ProdCountMap) {
	for prod, count := range g1 {
		pcm[prod] += count
	}
}

func (pcm ProdCountMap) Sub(g1 ProdCountMap) {
	for prod, count := range g1 {
		pcm[prod] -= count
		if pcm[prod] <= 0 {
			delete(pcm, prod)
		}
	}
}

func (pcm ProdCountMap) DeepCopy() ProdCountMap {
	cp := make(ProdCountMap)
	for k, v := range pcm {
		cp[k] = v
	}
	return cp
}

func (pcm ProdCountMap) TotalCount() int {
	totalCount := 0
	for _, count := range pcm {
		totalCount += count
	}
	return totalCount
}

// NUMA map[address]nodeID
type NUMA map[string]string
