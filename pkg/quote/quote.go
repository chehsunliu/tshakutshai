package quote

type Quote struct {
	Code string
	Name string

	Volume       int
	Value        int
	Transactions int
	Open         float32
	Close        float32
	High         float32
	Low          float32
}
