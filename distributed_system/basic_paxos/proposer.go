package basic_paxos

import "fmt"

type Proposer struct {
	// 服务器 id
	id int

	// 当前提议者已知最大轮次
	round int

	// 提案编号，它由服务器 id 和轮次 round 一起组成
	proposalNumber int

	// 接受者 id 列表
	acceptors []int
}

// 两阶段：
// 第一阶段 prepare 准备，
// 第二阶段 accept 接受
func (p *Proposer) propose(v any) any {
	p.round++ // 先将轮次自增
	p.proposalNumber = p.GetProposalNumber()

	// 第一阶段 ---- 准备阶段 prepare
	prepareCount := 0 // 收到正常回复的 acceptor 数量
	maxNumber := 0    // 收到的最大提案编号

	// 向大多数 acceptor 发送 prepare 消息，这个大多数指的是 acceptor 数量的一半 + 1。
	for _, aid := range p.acceptors {
		args := MsgArgs{
			Number: p.proposalNumber,
			From:   p.id,
			To:     aid,
		}
		reply := new(MsgReply)
		ok := call(fmt.Sprintf("127.0.0.1:%d", aid), "Acceptor.Prepare", args, reply)
		if !ok {
			continue
		}

		// 收到正常回复，更新收到回复的数量 prepareCount
		if reply.Ok {
			prepareCount++
			// 当回复的提案编号比当前的 maxNumber 大，更新 maxNumber 和 提案值 v。
			if reply.Number > maxNumber {
				maxNumber = reply.Number
				v = reply.Value
			}
		}

		// 收到了多数回应时，结束准备阶段，退出循环
		if prepareCount == p.majority() {
			break
		}
	}

	// 第二阶段 ---- 接受阶段 accept
	acceptCount := 0 // 成功接受提案的 acceptor 数量，同样是要求该数量大于等于接受者数量 / 2 + 1

	// 如果准备阶段的成功回复数量，已经满足多数派（接收者数量 / 2 + 1），则向所有接受者发送 accept 消息。
	if prepareCount >= p.majority() {
		for _, aid := range p.acceptors {
			args := MsgArgs{
				Number: p.proposalNumber, // 提案编号还是一开始生成的编号
				Value:  v,                // 提案值是 prepare 阶段收到的回复消息中，最大编号对应的提案值
				From:   p.id,
				To:     aid,
			}
			reply := new(MsgReply)
			ok := call(fmt.Sprintf("127.0.0.1:%d", aid), "Acceptor.Accept", args, reply)
			if !ok {
				continue
			}

			if reply.Ok {
				acceptCount++
			}
		}
	}

	if acceptCount >= p.majority() {
		return v
	}
	return nil
}

func (p *Proposer) majority() int {
	return len(p.acceptors)/2 + 1
}

func (p *Proposer) GetProposalNumber() int {
	return p.round<<16 | p.id
}
