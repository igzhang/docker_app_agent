package main

type ContainerConfigDetails struct {
	Cmd    []string `json:"cmd,omitempty"`
	Env    []string `json:"env,omitempty"`    // example: ENV=123
	Port   []string `json:"port,omitempty"`   // example: 6001:5001/udp
	Mount  []string `json:"mount,omitempty"`  // example: /host/path:/container/path
	Cpu    float32  `json:"cpu,omitempty"`    // C
	Memory float32  `json:"memory,omitempty"` // M
}

type ReceiveCMD struct {
	Operate      string                 `json:"operate"`
	Image        string                 `json:"image"`
	RegistryAuth string                 `json:"registry_auth,omitempty"`
	AppName      string                 `json:"app_name"`
	Extra        ContainerConfigDetails `json:"extra"`
}

type ResultCMD struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

const ResultSuccess = 1
const ResultFailure = 2
