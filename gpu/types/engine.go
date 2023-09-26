package types

// EngineParams .
type EngineParams struct {
	ProdCountMap ProdCountMap `json:"prod_count_map" mapstructure:"prod_count_map"`
}

func (ep *EngineParams) Count() int {
	return ep.ProdCountMap.TotalCount()
}
