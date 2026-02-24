//go:build !go1.24 || !cache

package govaluate

func getParameterStage(name string) (*evaluationStage, error) {
	operator := makeParameterStage(name)
	return &evaluationStage{
		operator: operator,
	}, nil
}

func getConstantStage(value any) (*evaluationStage, error) {
	operator := makeLiteralStage(value)
	return &evaluationStage{
		symbol:   LITERAL,
		operator: operator,
	}, nil
}
