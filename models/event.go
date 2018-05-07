package models

import (
	_ "github.com/go-sql-driver/mysql"
)

type Event struct {
	Id         int    `gorm:"primary_key;auto_increment;column:id"`
	EventName  string `gorm:"column:event_name"`
	EventType  string `gorm:"column:event_type"`
	OcTime     int    `gorm:"column:oc_time;type:int(11)"`
	DealStatus string `gorm:"column:deal_status"`
	DealTime   int    `gorm:"column:deal_time;type:int(11)"`
}

type EventDealRule struct {
	EventDealRuleId int    `gorm:"primary_key;auto_increment;column:event_deal_rule_id"`
	EventType       string `gorm:"column:event_type"`
	DealOrder       int    `gorm:"column:deal_order"`
	ActionId        int    `gorm:"column:action_id"`
	Note            string `gorm:"column:note"`
}

type EventDealInst struct {
	EventDealInstId int `gorm:"primary_key;auto_increment;column:event_deal_inst_id"`
	RefDealRuleId   int `gorm:"column:ref_deal_rule_id"`
	RefAppInstId    int `gorm:"column:ref_app_inst_id"`
	EventId         int `gorm:"column:event_id"`
}

type EvnetAction struct {
	ActionId   int    `gorm:"primary_key;auto_increment;column:action_id"`
	ActionType string `gorm:"column:action_type"`
	ActionItem string `gorm:"column:action_item"`
	Note       string `gorm:"column:note"`
}
