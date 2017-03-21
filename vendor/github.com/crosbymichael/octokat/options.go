package octokat

type Options struct {
	Headers     Headers
	Params      interface{}
	QueryParams map[string]string
}

type Headers map[string]string
