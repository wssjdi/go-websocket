package retcode

const (
	SUCCESS = 0  //请求成功
	FAIL    = -1 //请求出错

	SYSTEM_ID_ERROR      = -1001 //系统ID不能为空
	ONLINE_MESSAGE_CODE  = 1001  //客户端上线
	OFFLINE_MESSAGE_CODE = 1002  //客户端下线
)
