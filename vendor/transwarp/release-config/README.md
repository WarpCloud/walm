1. 选择 https://github.com/kubernetes/code-generator 合适的版本
2. export CODEGEN_PKG=../../k8s.io/code-generator/; ./hack/update-codegen.sh
3. 将对应code-generator的 Godeps/Godeps.json 替换项目本身的 Godeps/Godeps.json
