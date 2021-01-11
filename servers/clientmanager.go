package servers

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/woodylan/go-websocket/define/retcode"
	"github.com/woodylan/go-websocket/pkg/setting"
	"github.com/woodylan/go-websocket/tools/util"
	"strings"
	"sync"
	"time"
)

// 连接管理
type ClientManager struct {
	ClientIdMap     map[string]*Client // 全部的连接
	ClientIdMapLock sync.RWMutex       // 读写锁

	Connect    chan *Client // 连接处理
	DisConnect chan *Client // 断开连接处理

	GroupLock sync.RWMutex
	Groups    map[string][]string // 群组
	// key为GroupName;value为ClientId列表

	UserLock    sync.RWMutex
	UserClients map[string][]string // 拥有同样业务端UserId的客户端连接
	// 用户当一个业务端用户在多个地方登陆时通知客户端
	// key为业务端userId;value为ClientId列表

	SystemClientsLock sync.RWMutex
	SystemClients     map[string][]string // 所有系统的链接
	// key为systemId;value为ClientId列表
}

func NewClientManager() (clientManager *ClientManager) {
	clientManager = &ClientManager{
		ClientIdMap:   make(map[string]*Client),
		Connect:       make(chan *Client, 10000),
		DisConnect:    make(chan *Client, 10000),
		Groups:        make(map[string][]string, 100),
		UserClients:   make(map[string][]string, 100),
		SystemClients: make(map[string][]string, 100),
	}

	return
}

// 管道处理程序
func (manager *ClientManager) Start() {
	for {
		select {
		case client := <-manager.Connect:
			// 建立连接事件
			manager.EventConnect(client)
		case conn := <-manager.DisConnect:
			// 断开连接事件
			manager.EventDisconnect(conn)
		}
	}
}

// 建立连接事件
func (manager *ClientManager) EventConnect(client *Client) {
	manager.AddClient(client)

	log.WithFields(log.Fields{
		"host":     setting.GlobalSetting.LocalHost,
		"port":     setting.CommonSetting.HttpPort,
		"clientId": client.ClientId,
		"counts":   Manager.Count(),
	}).Info("客户端已连接")
}

// 断开连接时间
func (manager *ClientManager) EventDisconnect(client *Client) {
	//关闭连接
	_ = client.Socket.Close()
	manager.DelClient(client)

	mJson, _ := json.Marshal(map[string]string{
		"systemId":  client.SystemId,
		"groupName": strings.Join(client.GroupList, ","),
		"clientId":  client.ClientId,
		"userId":    client.UserId,
		"extend":    client.Extend,
	})
	data := string(mJson)

	//发送下线通知
	//通知同UserId的客户端连接
	if len(client.UserId) > 0 {
		//默认通知所有当前用户登录的客户端，不区分system和group
		SendMessage2User("", client.ClientId, "", client.UserId, retcode.OffLineMsgCode, "客户端下线", &data)
	}

	//通知同组的客户端连接
	if client.Notify && len(client.GroupList) > 0 {
		for _, groupName := range client.GroupList {
			SendMessage2Group(client.SystemId, client.ClientId, groupName, retcode.OffLineMsgCode, "客户端下线", &data)
		}
	}

	log.WithFields(log.Fields{
		"host":     setting.GlobalSetting.LocalHost,
		"port":     setting.CommonSetting.HttpPort,
		"clientId": client.ClientId,
		"counts":   Manager.Count(),
		"seconds":  uint64(time.Now().Unix()) - client.ConnectTime,
	}).Info("客户端已断开")

	//标记销毁
	client.IsDeleted = true
	client = nil
}

// 添加客户端
func (manager *ClientManager) AddClient(client *Client) {
	manager.ClientIdMapLock.Lock()
	defer manager.ClientIdMapLock.Unlock()

	manager.ClientIdMap[client.ClientId] = client
}

// 获取所有的客户端
func (manager *ClientManager) AllClient() map[string]*Client {
	manager.ClientIdMapLock.RLock()
	defer manager.ClientIdMapLock.RUnlock()

	return manager.ClientIdMap
}

// 客户端数量
func (manager *ClientManager) Count() int {
	manager.ClientIdMapLock.RLock()
	defer manager.ClientIdMapLock.RUnlock()
	return len(manager.ClientIdMap)
}

// 删除客户端
func (manager *ClientManager) DelClient(client *Client) {
	manager.delClientIdMap(client.ClientId)

	//删除用户自己列表
	if len(client.UserId) > 0 {
		manager.delUserClient(client.UserId, client.ClientId)
	}

	//删除所在的分组
	if len(client.GroupList) > 0 {
		for _, groupName := range client.GroupList {
			manager.delGroupClient(util.GenGroupKey(client.SystemId, groupName), client.ClientId)
		}
	}
	// 删除系统里的客户端
	manager.delSystemClient(client)
}

// 删除clientIdMap
func (manager *ClientManager) delClientIdMap(clientId string) {
	manager.ClientIdMapLock.Lock()
	defer manager.ClientIdMapLock.Unlock()

	delete(manager.ClientIdMap, clientId)
}

// 通过clientId获取
func (manager *ClientManager) GetByClientId(clientId string) (*Client, error) {
	manager.ClientIdMapLock.RLock()
	defer manager.ClientIdMapLock.RUnlock()

	if client, ok := manager.ClientIdMap[clientId]; !ok {
		return nil, errors.New("客户端不存在")
	} else {
		return client, nil
	}
}

// 发送到本机分组
func (manager *ClientManager) SendMessage2LocalGroup(systemId, messageId, sendUserId, groupName string, code int, msg string, data *string) {
	if len(groupName) > 0 {
		clientIds := manager.GetGroupClientList(util.GenGroupKey(systemId, groupName))
		if len(clientIds) > 0 {
			for _, clientId := range clientIds {
				if len(sendUserId) > 0 && sendUserId == clientId {
					continue
				}
				if _, err := Manager.GetByClientId(clientId); err == nil {
					//添加到本地
					SendMessage2LocalClient(messageId, clientId, sendUserId, code, msg, data)
				} else {
					//如果客户端连接已经不存在了,则从group中删除
					manager.delGroupClient(util.GenGroupKey(systemId, groupName), clientId)
				}
			}
		}
	}
}

// 发送到本机对应的userId
func (manager *ClientManager) SendMessage2LocalUserId(systemId, messageId, sendUserId, groupName, userId string, code int, msg string, data *string) {
	if len(userId) > 0 {
		userClients := manager.GetUserClients(userId)
		if len(userClients) > 0 {
			//log.Infof("SendMessage2LocalUserId userClients [ %d ]", len(userClients))
			for _, clientId := range userClients {
				//log.Infof("SendMessage2LocalUserId clientId [ %s ]", clientId)
				if len(sendUserId) > 0 && sendUserId == clientId {
					continue //是自己,不发消息给自己
				}

				var client *Client
				var err error
				if client, err = manager.GetByClientId(clientId); err != nil {
					continue //跳过,当前连接已经不存在
				}

				if client.IsDeleted {
					continue //跳过,当前连接已经关闭
				}

				//如果查询条件中传了systemId
				if len(systemId) > 0 && client.SystemId != systemId {
					continue //跳过,判断下一个连接
				}
				//log.Infof("SendMessage2LocalUserId systemId [ %s ]", systemId)

				send := true
				if len(groupName) > 0 {
					//log.Infof("SendMessage2LocalUserId groupName [ %s ]", groupName)
					send = false
					groupList := client.GroupList
					for _, myGroup := range groupList {
						if groupName == myGroup {
							send = true
						}
					}
				}

				//log.Infof("SendMessage2LocalUserId messageId [ %s ]", messageId)
				if send {
					SendMessage2LocalClient(messageId, clientId, sendUserId, code, msg, data)
				}
			}
		}
	}
}

//发送给指定业务系统
func (manager *ClientManager) SendMessage2LocalSystem(systemId, messageId string, sendUserId string, code int, msg string, data *string) {
	if len(systemId) > 0 {
		clientIds := Manager.GetSystemClientList(systemId)
		if len(clientIds) > 0 {
			for _, clientId := range clientIds {
				SendMessage2LocalClient(messageId, clientId, sendUserId, code, msg, data)
			}
		}
	}
}

// 添加到本地分组
func (manager *ClientManager) AddClient2LocalGroup(groupName string, client *Client, userId string, extend string) {
	//标记当前客户端的userId
	client.UserId = userId
	client.Extend = extend

	//判断之前是否有添加过
	for _, groupValue := range client.GroupList {
		if groupValue == groupName {
			return
		}
	}

	// 为属性添加分组信息
	groupKey := util.GenGroupKey(client.SystemId, groupName)

	manager.addClient2Group(groupKey, client)

	client.GroupList = append(client.GroupList, groupName)

	if len(userId) > 0 {
		//log.Info("ClientManager AddClient2LocalGroup userId:[%s], group:[%s], clientId:[%s]", userId, groupName, client.ClientId)
		manager.AddClient2UserClients(userId, groupName, client)
	}

	mJson, _ := json.Marshal(map[string]string{
		"systemId":  client.SystemId,
		"groupName": groupName,
		"clientId":  client.ClientId,
		"userId":    client.UserId,
		"extend":    client.Extend,
	})
	data := string(mJson)

	if client.Notify {
		//发送系统通知
		SendMessage2Group(client.SystemId, client.ClientId, groupName, retcode.OnLineMsgCode, "客户端上线", &data)
	}
}

// 添加到本地分组
func (manager *ClientManager) addClient2Group(groupKey string, client *Client) {
	manager.GroupLock.Lock()
	defer manager.GroupLock.Unlock()
	manager.Groups[groupKey] = append(manager.Groups[groupKey], client.ClientId)
}

// 删除分组里的客户端
func (manager *ClientManager) delGroupClient(groupKey string, clientId string) {
	manager.GroupLock.Lock()
	defer manager.GroupLock.Unlock()

	for index, groupClientId := range manager.Groups[groupKey] {
		if groupClientId == clientId {
			manager.Groups[groupKey] = append(manager.Groups[groupKey][:index], manager.Groups[groupKey][index+1:]...)
		}
	}
}

// 获取本地分组的成员
func (manager *ClientManager) GetGroupClientList(groupKey string) []string {
	manager.GroupLock.RLock()
	defer manager.GroupLock.RUnlock()
	return manager.Groups[groupKey]
}

// 添加到用户客户端连接列表
func (manager *ClientManager) AddClient2UserClients(userId, groupName string, client *Client) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	if len(userId) == 0 {
		return
	}

	//判断之前是否有添加过
	for _, clientId := range manager.UserClients[userId] {
		if clientId == client.ClientId {
			return
		}
	}

	manager.UserClients[userId] = append(manager.UserClients[userId], client.ClientId)

	mJson, _ := json.Marshal(map[string]string{
		"systemId":  client.SystemId,
		"groupName": groupName,
		"clientId":  client.ClientId,
		"userId":    userId,
		"extend":    client.Extend,
	})
	data := string(mJson)
	//默认通知所有当前用户登录的客户端，不区分system和group
	SendMessage2User("", client.ClientId, "", userId, retcode.MultiSignOnCode, "在另外一个客户端登录", &data)
}

// 删除用户列表里的客户端连接
func (manager *ClientManager) delUserClient(userId string, clientId string) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	userClients := manager.UserClients[userId]
	//只有一个并且就是要删除的这个连接
	if len(userClients) == 1 && clientId == userClients[0] {
		delete(manager.UserClients, userId)
	} else {
		for index, userClientId := range userClients {
			if userClientId == clientId {
				manager.UserClients[userId] = append(userClients[:index], userClients[index+1:]...)
			}
		}
	}
}

// 获取本地用户列表里的客户端连接clientId
func (manager *ClientManager) GetUserClients(userId string) []string {
	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()
	return manager.UserClients[userId]
}

// 获取本地用户列表里的客户端连接,返回内容格式为:[systemId:groupName:clientId]
func (manager *ClientManager) GetSystemGroupUserClients(systemId, groupName, userId string) []string {
	var userClients []string
	for _, userClientId := range manager.GetUserClients(userId) {
		var client *Client
		var err error
		if client, err = manager.GetByClientId(userClientId); err != nil {
			continue //跳过,当前连接已经不存在
		}

		if client.IsDeleted {
			continue //跳过,当前连接已经关闭
		}

		//如果查询条件中传了systemId
		if len(systemId) > 0 && client.SystemId != systemId {
			continue //跳过,判断下一个连接
		}

		groupList := client.GroupList
		if len(groupList) == 0 {
			if len(groupName) > 0 {
				continue //跳过,判断下一个组
			}
			userClients = append(userClients, util.GenUserClientKey(systemId, "", userClientId))
		} else {
			for _, myGroup := range client.GroupList {
				if len(groupName) > 0 && groupName != myGroup {
					continue //跳过,判断下一个组
				}
				userClients = append(userClients, util.GenUserClientKey(systemId, myGroup, userClientId))
			}
		}
	}

	return userClients
}

// 添加到系统客户端列表
func (manager *ClientManager) AddClient2SystemClient(systemId string, client *Client) {
	manager.SystemClientsLock.Lock()
	defer manager.SystemClientsLock.Unlock()
	manager.SystemClients[systemId] = append(manager.SystemClients[systemId], client.ClientId)
}

// 删除系统里的客户端
func (manager *ClientManager) delSystemClient(client *Client) {
	manager.SystemClientsLock.Lock()
	defer manager.SystemClientsLock.Unlock()

	for index, clientId := range manager.SystemClients[client.SystemId] {
		if clientId == client.ClientId {
			manager.SystemClients[client.SystemId] = append(manager.SystemClients[client.SystemId][:index], manager.SystemClients[client.SystemId][index+1:]...)
		}
	}
}

// 获取指定系统的客户端列表
func (manager *ClientManager) GetSystemClientList(systemId string) []string {
	manager.SystemClientsLock.RLock()
	defer manager.SystemClientsLock.RUnlock()
	return manager.SystemClients[systemId]
}
