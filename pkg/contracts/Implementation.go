package contracts

type ResponseImplementation struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             interface{}
}
