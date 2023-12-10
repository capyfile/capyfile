#!/bin/bash

# Check if an app name is provided as the first argument
if [ $# -lt 1 ]; then
  echo "Usage: $0 <app_name> [start|stop|shell]"
  exit 1
fi

# Define the app name and Docker Compose file
app_name=$1

# Check if the app_name is valid and perform the requested action
case $app_name in
  "capysvr" | "capycmd" | "capyworker")
    echo "Performing action on $app_name..."
    ;;
  *)
    echo "Unknown app: $app_name"
    exit 1
    ;;
esac

# We need to run different services depending on the app.
if [ "$app_name" = "capysvr" ]; then
  docker_compose_services="proxy minio minio-buckets"
elif [ "$app_name" = "capycmd" ] || [ "$app_name" = "capyworker" ]; then
  docker_compose_services="proxy minio minio-buckets"
else
  echo "Unknown app: $app_name"
  exit 1
fi

docker_compose_file="docker-compose.dev.yml"

# Function to start the Docker Compose service
start_app() {
  if [ "$1" = "capysvr" ]; then
    docker compose -f $docker_compose_file up -d $docker_compose_services $1
  elif [ "$1" = "capycmd" || "$1" = "capyworker" ]; then
    docker compose -f $docker_compose_file up -d $docker_compose_services; \
      docker compose -f $docker_compose_file run --entrypoint "" $1 sh
  else
    echo "Unknown app: $1"
    exit 1
  fi
}

# Function to stop the Docker Compose service
stop_app() {
  docker compose -f $docker_compose_file down
}

# Function to rebuild the Docker Compose service
rebuild_app() {
  docker compose -f $docker_compose_file build $1
}

# Function to shell into the Docker container
shell_into_container() {
  docker compose -f $docker_compose_file exec $1 sh
}

# Check if the action is to start, stop, or shell into the app
action="$2"
if [ "$action" = "start" ]; then
  start_app $app_name
elif [ "$action" = "stop" ]; then
  stop_app
elif [ "$action" = "rebuild" ]; then
  rebuild_app $app_name
elif [ "$action" = "shell" ]; then
  shell_into_container $app_name
else
  echo "Usage: $0 <app_name> [start|stop|rebuild|shell]"
  exit 1
fi
