package config

import "net/url"

type MockConfiguration struct {
}

func (c *MockConfiguration) GetArmoryCloudEnv() ArmoryCloudEnv {
	return 0
}
func (c *MockConfiguration) GetAuthToken() string {
	return "abc_xyz_token"
}
func (c *MockConfiguration) GetCustomerEnvironmentId() string {
	return ""
}
func (c *MockConfiguration) GetArmoryCloudAddr() *url.URL {
	addr, _ := url.Parse("api.dev.cloud.armory.io")
	return addr
}
func (c *MockConfiguration) GetArmoryCloudEnvironmentConfiguration() *ArmoryCloudEnvironmentConfiguration {
	return &ArmoryCloudEnvironmentConfiguration{
		CloudConsoleBaseUrl: "console.dev.cloud.armory.io",
	}
}
