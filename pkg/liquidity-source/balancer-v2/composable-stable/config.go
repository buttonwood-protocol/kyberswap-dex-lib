package composablestable

type Config struct {
	DexID        string `json:"dexID"`
	SubgraphAPI  string `json:"subgraphAPI"`
	NewPoolLimit int    `json:"newPoolLimit"`
}
