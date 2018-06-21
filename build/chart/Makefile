CHART.Name:=walm
CHART.Version:=1.2
CHART.Repo:=http://172.16.1.41:8080/api/chart

RUN_IN_DOCKER:=docker run -it -v `pwd`:/root/$(CHART.Name)  -w /root/$(CHART.Name)  172.16.1.99/gold/helm:tos18-latest

default:
	$(RUN_IN_DOCKER) helm package -u walm
	@echo "package chart ok"

pull:
	$(RUN_IN_DOCKER) helm fetch $(CHART.Name):$(CHART.Version)
	@echo "pull ok"

push:
	curl --data-binary "$(CHART.Name):$(CHART.Version).tgz" $(CHART.Repo)
	@echo "push ok"

create:
	helm install --name walm-helm-test --link txsql= walm 

clean:
	rm -f walm-*.tgz
	@echo "clean ok"
