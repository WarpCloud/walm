#!/bin/bash

confd -onetime -backend file -prefix / -file /etc/confd/txsql-confd.conf
/bin/boot.sh TXSQL_SERVER