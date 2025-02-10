build:
	docker build -t todoserv -f Dockerfile.be backend
	docker build -t todofront -f Dockerfile.fe frontend
stop:
	docker stop todoserv 2>/dev/null || true
	docker stop todofront 2>/dev/null || true

run: stop
	docker run -d --rm -p 2000:2000 --name todoserv todoserv:latest /bin/todoserv -db /db.bin
	docker run -d --rm -p 8080:3000 --name todofront todofront:latest

