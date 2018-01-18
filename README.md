# WALM

Warp Application Lifecycle Manager

# Development

Python 3.5+

## 初始化

初始化过程进行一些基本的数据库操作，如导入模板数据。

```bash
git clone --recursive ssh://git@172.16.1.41:10022/TDC/walm.git
python setup.py install
walm db upgrade
```

创建数据库

```mysql
CREATE DATABASE IF NOT EXISTS walmdb DEFAULT CHARSET utf8 COLLATE utf8_general_ci;
```

## Run

```bash
export PRODUCT_META_HOME=/path/to/product-meta
export KUBERNETES_HOST=172.16.3.234:8080
export SQLALCHEMY_DATABASE_URI=mysql://root:password@localhost:3306/walmdb?charset=utf8

walm runserver --debug
```