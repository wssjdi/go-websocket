package retcode

const (
	//错误响应码都 < 0
	SystemIdErrCode = -1001 //系统ID无效
	FAIL            = -1    //请求出错

	//成功响应码都 >= 0
	SUCCESS         = 0    //请求成功
	OnLineMsgCode   = 1001 //客户端上线
	OffLineMsgCode  = 1002 //客户端下线
	MultiSignOnCode = 1003 //业务端同意用户多点登录通知
)
