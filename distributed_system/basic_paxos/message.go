package basic_paxos

import "net/rpc"

type MsgArgs struct {
	// 提案编号
	Number int

	// 提案值
	Value any

	// 发送者 id
	From int

	// 接受者 id
	To int
}

type MsgReply struct {
	Ok     bool
	Number int
	Value  any
}

// call 用于发送 RPC 消息并接收响应。
func call(srv string, name string, args any, reply any) bool {
	c, err := rpc.Dial("tcp", srv)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(name, args, reply)
	if err != nil {
		return false
	}
	return true
}
