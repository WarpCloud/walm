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

## Run

```bash
export PRODUCT_META_HOME=/path/to/product-meta

walm runserver --debug
```