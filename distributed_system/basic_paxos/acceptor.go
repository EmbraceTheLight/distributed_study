package basic_paxos

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Acceptor struct {
	lis net.Listener

	// 接受者 id, 存储接受者服务的 RPC 端口
	id int

	// 第一阶段接受者承诺的提案编号。为 0 则表示接受者没有收到过任何提案者发送的 Prepare 消息
	minProposalNumber int

	// 接受者已接受的提案编号。如果为 0，表示没有接受任何提案
	acceptedNumber int

	// 接受者已接受的提案值，如果没有接受任何提案，则为 nil
	acceptedValue any

	// 学习者 id 列表, 存储学习者 RPC 端口
	learners []int
}

func (a *Acceptor) Prepare(args *MsgArgs, reply *MsgReply) error {
	if args.Number > a.minProposalNumber {
		// 接受者收到的 prepare 消息编号大于已接受的编号，则更新接收的提案编号，表示不会再接受任何编号小于该值的提案
		a.minProposalNumber = args.Number
		// 返回已接受的编号和值
		reply.Number = a.acceptedNumber
		reply.Value = a.acceptedValue
		reply.Ok = true
	} else { // 否则，返回 false，表示已接受了更大的编号的提案
		reply.Ok = false
	}
	return nil
}

func (a *Acceptor) Accept(args *MsgArgs, reply *MsgReply) error {
	// 由提议者 proposer 传递过来的提案值不小于第一阶段接受者承诺的提案编号，则接受该提案，并向所有学习者发送学习消息
	if args.Number >= a.minProposalNumber {
		a.minProposalNumber = args.Number
		a.acceptedNumber = args.Number
		a.acceptedValue = args.Value
		reply.Ok = true

		for _, lid := range a.learners {
			// 开启协程，向学习者发送学习消息
			go func(learnerPort int) {
				// 注意修改 From 和 To 字段，以便于找到学习者
				addr := fmt.Sprintf("127.0.0.1:%d", learnerPort)
				args.From = a.id
				args.To = learnerPort
				resp := new(MsgReply)
				ok := call(addr, "Learner.Learn", args, resp)
				if !ok {
					return
				}
			}(lid)
		}
	} else {
		reply.Ok = false
	}
	return nil
}

func newAcceptor(id int, learners []int) *Acceptor {
	acceptor := &Acceptor{
		id:       id,
		learners: learners,
	}
	acceptor.server()
	return acceptor
}

func (a *Acceptor) server() {
	rpcs := rpc.NewServer()
	rpcs.Register(a)
	addr := fmt.Sprintf(":%d", a.id)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen error: ", err)
	}
	a.lis = lis

	go func() {
		for {
			conn, err := a.lis.Accept()
			if err != nil {
				continue
			}
			go rpcs.ServeConn(conn)
		}
	}()
}

// 关闭连接
func (a *Acceptor) close() {
	a.lis.Close()
}
