package types

type StepExecutionResult struct {
	IsFinal  bool
	NextStep *string
}
