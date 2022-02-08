package protocol

// ReqSignUp 注册请求.
type ReqSignUp struct {
	UserName string `json:"user_name"` // 用户名, 不为空
	Password string `json:"password"`  // 密码, 不为空
	NickName string `json:"nick_name"` // 昵称
}
type RespSignUp struct {
	Ret int `json:"ret"` // 结果码 0:成功 1:用户名或密码为空 2:用户名重复或创建失败
}

type ReqGetProfile struct {
	UserName string
	Token    string
}
type RespGetProfile struct {
	Ret      int    `json:"ret"`       // 结果码 0:成功 1:token校验失败 2:数据为空 3:获取失败
	UserName string `json:"user_name"` // 用户名，不为空
	NickName string `json:"nick_name"` // 昵称
	PicName  string `json:"pic_name"`  // 头像(路径信息)
}

// ReqLogin 登录请求.
type ReqLogin struct {
	UserName string `json:"user_name"` // 用户名, 不为空
	Password string `json:"password"`  // 密码, 不为空
}

// RespLogin 登录返回.
type RespLogin struct {
	Ret   int    `json:"ret"`   // 结果码 0:成功 1:用户名或密码错误 2:登录失败
	Token string `json:"token"` // token
}
