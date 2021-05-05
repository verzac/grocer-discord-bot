registry_uri = public.ecr.aws/t5m8k1a3
image_name = grocer-discord-bot
repository = ${image_name}
repository_uri = ${registry_uri}/${repository}
tag = ${shell git describe --tags}

docker_login:
	aws ecr-public get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(registry_uri)

docker_build:
	docker build -t ${image_name}:latest -t ${image_name}:${tag} --build-arg version=$(tag) .

publish: docker_build docker_login
	docker tag ${image_name}:latest ${repository_uri}:latest
	docker tag ${image_name}:${tag} ${repository_uri}:${tag}
	docker push ${repository_uri}:latest
	docker push ${repository_uri}:${tag}