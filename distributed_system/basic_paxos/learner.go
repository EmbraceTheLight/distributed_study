package basic_paxos

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Learner struct {
	lis net.Listener

	// 学习者 id，使用端口号作为其 id
	id int

	// 记录接受者已接受的提案: key 为接受者 id，value 为请求消息
	acceptedMsg map[int]MsgArgs
}

func (learner *Learner) Learn(args *MsgArgs, reply *MsgReply) error {
	accepted := learner.acceptedMsg[args.From]

	// 如果接受者的提案编号小于当前学习者收到的提案的编号，更新该接受者提案信息
	if accepted.Number < args.Number {
		learner.acceptedMsg[args.From] = *args
		reply.Ok = true
	} else {
		reply.Ok = false
	}
	return nil
}

func (learner *Learner) chosen() any {
	acceptCounts := make(map[int]int)  // 键：提案编号，值：接受该提案的节点数
	acceptMsg := make(map[int]MsgArgs) // 键：提案编号，值：该提案的消息
	for _, accepted := range learner.acceptedMsg {
		// 提案编号不为 0，表示接受者已经接受过提案
		if accepted.Number != 0 {
			acceptCounts[accepted.Number]++
			acceptMsg[accepted.Number] = accepted
		}
	}

	// 遍历被接受的提案，找出已被大多数节点接受的提案，并返回该提案的值。
	for n, count := range acceptCounts {
		if count >= learner.majority() {
			return acceptMsg[n].Value
		}
	}
	return nil
}

func (learner *Learner) majority() int {
	return len(learner.acceptedMsg)/2 + 1
}

func (learner *Learner) newLearner(id int, acceptorIds []int) *Learner {
	l := &Learner{
		id:          id,
		acceptedMsg: make(map[int]MsgArgs),
	}

	// 初始化接受者提案 map 信息
	for _, aid := range acceptorIds {
		l.acceptedMsg[aid] = MsgArgs{
			Number: 0,
			Value:  nil,
		}
	}
	learner.server(id)
	return learner
}

func (learner *Learner) server(id int) {
	rpcs := rpc.NewServer()
	rpcs.Register(learner)
	addr := fmt.Sprintf(":%d", id)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen error: ", err)
	}
	learner.lis = lis
	go func() {
		for {
			conn, err := learner.lis.Accept()
			if err != nil {
				continue
			}
			go rpcs.ServeConn(conn)
		}
	}()
}

// 关闭连接
func (learner *Learner) close() {
	learner.lis.Close()
}
