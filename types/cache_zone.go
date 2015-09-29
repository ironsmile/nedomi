package types

// CacheZone is the combination of a Storage for storing object parts and an
// `CacheAlgorithm` which determines what should be stored.
type CacheZone struct {
	ID        string
	PartSize  BytesSize
	Algorithm CacheAlgorithm
	Scheduler Scheduler
	Storage   Storage
}
