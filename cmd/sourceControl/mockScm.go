package sourceControl

type MockSmc struct {
	BaseScmc
}

func (mock MockSmc) GetContext() (ScmContext, error) {
	scmc := MockSmc{
		BaseScmc: BaseScmc{
			Type: "mock"}}
	return scmc, nil
}

func GetEmptyMockScmc() ScmContext {
	scmc, _ := MockSmc{}.GetContext()
	return scmc
}
