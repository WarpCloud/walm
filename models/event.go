package models

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Event struct {
	gorm.Model
	EventName  string `gorm:"column:event_name"`
	EventType  string `gorm:"column:event_type"`
	DealStatus string `gorm:"column:deal_status"`
}

func (Event) TableName() string {
	return "event"
}

type EventDealRule struct {
	gorm.Model
	EventDealRuleId int    `gorm:"primary_key;auto_increment;column:event_deal_rule_id"`
	EventType       string `gorm:"column:event_type"`
	DealOrder       int    `gorm:"column:deal_order"`
	ActionId        int    `gorm:"column:action_id"`
	Note            string `gorm:"column:note"`
}

func (EventDealRule) TableName() string {
	return "event_deal_rule"
}

type EventDealInst struct {
	gorm.Model
	EventDealInstId int `gorm:"primary_key;auto_increment;column:event_deal_inst_id"`
	RefDealRuleId   int `gorm:"column:ref_deal_rule_id"`
	RefAppInstId    int `gorm:"column:ref_app_inst_id"`
	EventId         int `gorm:"column:event_id"`
}

func (EventDealInst) TableName() string {
	return "event_deal_inst"
}

type EvnetAction struct {
	gorm.Model
	ActionId   int    `gorm:"primary_key;auto_increment;column:action_id"`
	ActionType string `gorm:"column:action_type"`
	ActionItem string `gorm:"column:action_item"`
	Note       string `gorm:"column:note"`
}

func (EvnetAction) TableName() string {
	return "event_action"
}
