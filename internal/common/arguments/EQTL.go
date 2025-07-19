package arguments

type EQTL struct {
	*Expression
	*QTLSpecific
	Regulatory     bool
	WithMutation   bool
	WithExpression bool
}
