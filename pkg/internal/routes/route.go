package routes

type Service struct {
	ServiceName     string
	ServiceStruct   string
	ServiceRegister string
	HandlerRegister string
}

var Services = []Service{
	{
		ServiceName:     "Auth",
		ServiceStruct:   "AuthServiceServer",
		ServiceRegister: "RegisterAuthServer",
		HandlerRegister: "RegisterAuthHandler",
	},
}
