package retcode

const (
	//错误响应码都 < 0
	ETcdErrCode     = -1002 //ETcd服务器错误
	SystemIdErrCode = -1001 //系统ID无效
	FAIL            = -1    //请求出错

	//成功响应码都 >= 0
	SUCCESS        = 0    //请求成功
	OnLineMsgCode  = 1001 //客户端上线
	OffLineMsgCode = 1002 //客户端下线

	MultiSignOnCode = 2000 //业务端同意用户多点登录通知
)
