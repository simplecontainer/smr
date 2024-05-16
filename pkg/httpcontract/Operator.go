package httpcontract

type ResponseOperator struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             map[string]any
}
