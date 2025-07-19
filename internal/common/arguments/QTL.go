package arguments

type QTL struct {
	*Common
	*QTLSpecific
}

type QTLSpecific struct {
	FreqIncrease bool
	Correction   bool
	FuncScore    bool

	MutRateParam       float64
	MutationDataFile   string
	FreqCutoff         float64
	WithinCondition    bool
	OutlierPopulations string
}
