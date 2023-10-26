package types

// EngineParams .
type EngineParams struct {
	ProdCountMap ProdCountMap `json:"prod_count_map" mapstructure:"prod_count_map"`
}

func (ep *EngineParams) Count() int {
	return ep.ProdCountMap.TotalCount()
}

func (ep *EngineParams) DeepCopy() *EngineParams {
	return &EngineParams{
		ProdCountMap: ep.ProdCountMap.DeepCopy(),
	}
}

func (ep *EngineParams) Sub(ep1 *EngineParams) {
	ep.ProdCountMap.Sub(ep1.ProdCountMap)
}

func (ep *EngineParams) Add(ep1 *EngineParams) {
	ep.ProdCountMap.Add(ep1.ProdCountMap)
}
