package basic_paxos

import (
	"fmt"
	"testing"
)

func start(acceptorIds []int, learnerIds []int) ([]*Acceptor, []*Learner) {
	acceptors := make([]*Acceptor, 0)
	for _, aid := range acceptorIds {
		acceptor := newAcceptor(aid, learnerIds)
		acceptors = append(acceptors, acceptor)
	}

	learners := make([]*Learner, 0)
	for _, lid := range learnerIds {
		learner := newLearner(lid, acceptorIds)
		learners = append(learners, learner)
	}

	return acceptors, learners
}

func cleanup(acceptors []*Acceptor, learners []*Learner) {
	for _, acceptor := range acceptors {
		acceptor.close()
	}

	for _, learner := range learners {
		learner.close()
	}
}

func TestSingleProposer(t *testing.T) {
	acceptorIds := []int{10001, 10002, 10003}
	learnerIds := []int{20001}
	acceptors, learners := start(acceptorIds, learnerIds)
	defer cleanup(acceptors, learners)

	p := &Proposer{
		id:        1,
		acceptors: acceptorIds,
	}

	value := p.propose("hello world")
	if value != "hello world" {
		t.Errorf("value = %s, expected %s", value, "hello world")
	}

	learnValue := learners[0].chosen()
	if learnValue != value {
		t.Errorf("learnValue = %s, expected %s", learnValue, "hello world")
	}
}

func TestTwoProposers(t *testing.T) {
	acceptorIds := []int{10001, 10002, 10003}
	learnerIds := []int{20001}
	acceptors, learners := start(acceptorIds, learnerIds)
	defer cleanup(acceptors, learners)

	p1 := &Proposer{
		id:        1,
		acceptors: acceptorIds,
	}
	v1 := p1.propose("hello world")

	p2 := &Proposer{
		id:        2,
		acceptors: acceptorIds,
	}
	v2 := p2.propose("hello book")
	if v1 != v2 {
		t.Errorf("value1 = %s, value2 = %s", v1, v2)
	}

	learnValue := learners[0].chosen()
	fmt.Printf("learn value is \"%s\"", learnValue)
	if learnValue != v1 {
		t.Errorf("learnValue = %s, expected %s", learnValue, v1)
	}
}
