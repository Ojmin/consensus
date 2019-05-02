package core

// 持币者
type Coiner struct {
	Address   string
	Amount    float64
	Candidate string
	IsVote    bool
}
