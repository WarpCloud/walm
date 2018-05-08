default:
	docker build -t walm-builder:1.0 .
	docker tag walm-builder:1.0 172.16.1.99/transwarp/walm-builder:1.0
	@echo "build ok"

pull:
	docker pull 172.16.1.99/transwarp/walm-builder:1.0
	@echo "pull ok"

push:
	docker push 172.16.1.99/transwarp/walm-builder:1.0
	@echo "push ok"

run:
	docker run --rm -it -v `pwd`:/root 172.16.1.99/transwarp/walm-builder:1.0

clean:
	@echo "clean ok"
