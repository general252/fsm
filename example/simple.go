package main

import (
	"github.com/general252/fsm"
	"log"
)

// 订单状态定义
const (
	StatePending   fsm.State = "待支付"
	StatePaid      fsm.State = "已支付"
	StateRefunding fsm.State = "退款中"
	StateShipped   fsm.State = "已发货"
	StateReturning fsm.State = "退货中"
	StateCanceled  fsm.State = "已取消"
	StateComplete  fsm.State = "已完成"
)

// 事件定义
const (
	EventPaySuccess     fsm.Event = "支付成功"
	EventApplyRefund    fsm.Event = "申请退款"
	EventRefuseRefund   fsm.Event = "拒绝退款"
	EventRefundSuccess  fsm.Event = "退款成功"
	EventCancel         fsm.Event = "取消订单"
	EventShip           fsm.Event = "发货"
	EventApplyReturn    fsm.Event = "申请退货"
	EventRefuseReturn   fsm.Event = "拒绝退货"
	EventReturnSuccess  fsm.Event = "退货成功"
	EventConfirmReceipt fsm.Event = "确认收货"
)

type eTo struct {
	F fsm.State
	E fsm.Event
	T fsm.State
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fm := fsm.NewStateMachine(StatePending)

	var eTos = []eTo{
		{StatePending, EventCancel, StateCanceled},
		{StatePending, EventPaySuccess, StatePaid},
		{StatePaid, EventCancel, StateRefunding},
		{StatePaid, EventApplyRefund, StateRefunding},
		{StateRefunding, EventRefundSuccess, StateCanceled},
		{StateRefunding, EventRefuseRefund, StatePaid},
		{StatePaid, EventShip, StateShipped},
		{StateShipped, EventApplyReturn, StateReturning},
		{StateReturning, EventRefuseReturn, StateShipped},
		{StateReturning, EventReturnSuccess, StateRefunding},
		{StateShipped, EventConfirmReceipt, StateComplete},
		{StateShipped, EventCancel, StateReturning},
		{StateReturning, EventConfirmReceipt, StateComplete},
	}

	for _, e := range eTos {
		err := fm.AddTransitions(&fsm.Transition{
			From:  e.F,
			Event: e.E,
			To:    e.T,
			Handle: func(from fsm.State, e fsm.Event, to fsm.State) error {
				log.Printf("处理事件... [%v]->(%v)->[%v]", from, e, to)
				return nil
			},
		})
		if err != nil {
			log.Println(err)
			return
		}
	}

	// 执行事件测试
	log.Println("当前状态:", fm.CurrentState())
	_ = fm.Trigger(EventPaySuccess) // 支付成功
	log.Println("当前状态:", fm.CurrentState())

	_ = fm.Trigger(EventShip) // 发货
	log.Println("当前状态:", fm.CurrentState())

	_ = fm.Trigger(EventCancel) // 取消订单
	log.Println("当前状态:", fm.CurrentState())

	// 测试非法转移
	err := fm.Trigger(EventApplyReturn)
	log.Println("尝试:", err) // 非法操作

	_, _, diagram := fm.View()
	log.Println("\n" + diagram)
}
