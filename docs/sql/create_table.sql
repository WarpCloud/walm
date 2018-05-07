CREATE DATABASE IF NOT EXISTS walmdb DEFAULT CHARSET utf8 COLLATE utf8_general_ci;

drop TABLE if EXISTS `product`;

CREATE TABLE `product`(
prod_id int UNSIGNED AUTO_INCREMENT comment "产品ID",
name VARCHAR(64) comment "产品名称",
note TEXT comment "说明",
app_list_id int UNSIGNED comment "应用列表ID",
vers VARCHAR(64) comment "产品版本",
config_temp TEXT comment "配置实例",
PRIMARY KEY(prod_id))ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "产品定义表";

--alter TABLE `product` add constraint `fk_product_app` foreign key(`app_list_id`)　REFERENCES `app_list`(`app_list_id`) ON DELETE SET NULL;

drop TABLE if EXISTS `app_list`;

create table IF NOT EXISTS `app_list`(
id int UNSIGNED NOT NULL AUTO_INCREMENT  comment "应用ID",
app_list_id int UNSIGNED  comment "应用列表ID",
app_pkg VARCHAR(64) NOT NULL comment "应用软件包",
vers VARCHAR(64) NOT NULL comment "应用软件包版本",
config_temp TEXT comment "配置实例",
PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "应用列表";

DROP TABLE IF EXISTS `app_dep_list`;

CREATE TABLE IF NOT EXISTS `app_dep_list`(
id int UNSIGNED NOT NULL AUTO_INCREMENT comment "应用依赖列表ID",
app_list_id int UNSIGNED NOT NULL comment "应用ID",
app_pkg VARCHAR(64) NOT NULL comment "应用依赖软件包",
vers VARCHAR(64) NOT NULL comment "应用依赖软件包版本",
PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "应用依赖关系表";

--alter TABLE `app_dep_list` add constraint `fk_app_dep` foreign key(`app_list_id`)　REFERENCES `app_list`(`id`) ON DELETE SET NULL;

DROP TABLE IF EXISTS `cluster`;

CREATE TABLE IF NOT EXISTS `cluster`(
cluster_id int UNSIGNED NOT NULL AUTO_INCREMENT comment "集群ID",
prod_id int UNSIGNED comment "产品ID",
config_temp TEXT comment "配置实例",
PRIMARY KEY(cluster_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "产品集群表";

--alter TABLE `cluster` add constraint `fk_app_cluster` foreign key(`prod_id`)　REFERENCES `product`(`prod_id`) ON DELETE SET NULL;

DROP TABLE IF EXISTS `cluster_app_ref_inst`;

CREATE TABLE IF NOT EXISTS `cluster_app_ref_inst`(
cluster_app_ref_inst_id int UNSIGNED NOT NULL AUTO_INCREMENT comment "集权应用引用关系ID",
cluster_id int UNSIGNED  NOT NULL  comment "集群ID",
app_inst_id int UNSIGNED NOT NULL  comment "应用实例ID",
PRIMARY KEY(cluster_app_ref_inst_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "集群实例对应关系表";

DROP TABLE IF EXISTS `app_inst`;

CREATE TABLE IF NOT EXISTS `app_inst`(
app_inst_id int UNSIGNED NOT NULL AUTO_INCREMENT comment "应用实例ID",
name VARCHAR(64) comment "应用实例名称",
app_pkg VARCHAR(64) NOT NULL comment "应用包",
vers VARCHAR(64)  comment "应用包版本",
config_temp TEXT  comment "配置模板",
status VARCHAR(64)  comment "应用实例状态",
install_time int(11)  comment "开始安装时间",
installed_time int(11)  comment "安装完成时间",
last_time int(11) comment "状态最后更新时间",
cluster_id int UNSIGNED comment "所属集群ID",
PRIMARY KEY(app_inst_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "应用实例表";

DROP TABLE IF EXISTS `app_inst_dep_list`;

CREATE TABLE IF NOT EXISTS `app_inst_dep_list`(
id int UNSIGNED NOT NULL AUTO_INCREMENT comment "标识ID",
app_inst_id int UNSIGNED NOT NULL comment "应用实例ID",
dep_inst_id int UNSIGNED NOT NULL comment "依赖应用实例ID",
PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "应用实例依赖表";

--alter TABLE `app_inst_dep_list` add constraint `fk_app_inst_dep_1` foreign key(`app_inst_id`)　REFERENCES `app_inst`(`app_inst_id`) ON DELETE SET NULL;


DROP TABLE IF EXISTS `event_deal_rule`;

CREATE TABLE if NOT EXISTS `event_deal_rule`(
event_deal_rule_id int UNSIGNED NOT NULL AUTO_INCREMENT comment "处理规则ID",
event_type VARCHAR(64)comment "事件类型",
deal_order int UNSIGNED comment "处理顺序",
action_id int UNSIGNED comment "动作ID",
note TEXT comment "注释",
PRIMARY KEY(event_deal_rule_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "事件处理规则表";

DROP TABLE IF EXISTS `event_deal_inst`;

CREATE TABLE IF NOT EXISTS `event_deal_inst`(
event_deal_inst_id int UNSIGNED NOT NULL AUTO_INCREMENT comment "事件处理实例ID",
ref_deal_rule_id int UNSIGNED comment "事件处理规则ID",
ref_app_inst_id int UNSIGNED comment "关联应用实例ID",
event_id int UNSIGNED comment "关联事件ID",
PRIMARY KEY(event_deal_inst)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "事件处理实例表";

DROP TABLE IF EXISTS  `evnet_action`;

CREATE TABLE IF NOT EXISTS `evnet_action`(
action_id int UNSIGNED NOT NULL AUTO_INCREMENT comment "动作ID",
action_type  VARCHAR(64) comment "动作类型",
action_item   VARCHAR(256) comment "具体动作",
note TEXT comment "注释",
PRIMARY KEY(action_id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "事件处理动作表";


DROP TABLE IF EXISTS `event`;

CREATE TABLE IF NOT EXISTS `event`(
id int UNSIGNED NOT NULL AUTO_INCREMENT comment "事件ID",
event_name VARCHAR(64)comment "事件名称",
event_type VARCHAR(64)comment "事件类型",
oc_time int(11) comment "发生时间",
deal_status  VARCHAR(64) comment"事件处理状态",
deal_time int(11) comment "事件处理时间",
PRIMARY KEY(id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 comment "事件实例表";
