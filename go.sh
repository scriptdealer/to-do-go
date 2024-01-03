#!/bin/bash -e

REPO_NAME=auth
COVERAGE_MIN=95

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

unit() {
  echo "running unit tests with -race"
   if ! go test -race ./... -count=1;
   then
     echo -e "${RED}[UNIT TESTS FAILED]${NC}"
     print_fail
     return 1
   else
     echo -e "${GREEN}[UNIT TESTS PASSED]${NC}"
   fi
}

unit_docker(){
  go mod vendor
  docker run --rm -v "$(pwd)":/work -w /work -it golang:1.21 \
    bash -c "git config --global --add safe.directory /work && ./run.sh unit"
  rm -Rf vendor
}

lint(){
  echo "run linter"
  go mod vendor
  if ! docker run --rm -v "$(pwd)":/work:ro -w /work -it golangci/golangci-lint:latest golangci-lint run -v
  then
    echo -e "${RED}[LINTER CHECK FAILED]${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}[LINTER CHECK PASSED]${NC}"
  fi
  rm -Rf vendor
}

deps_check() {
  echo "run dependency analyzer"
  go list -json -deps ./app | docker run -i --rm \
      sonatypecommunity/nancy:v1-alpine nancy sleuth
}

start_local() {
  echo "starting local services"

  # cd local
  docker compose pull
  docker compose build
  docker compose up -d
}

stop_local() {
  echo "stopping local services"

  # cd local
  docker compose down
}

start_integration() {
  echo "starting integration tests services"

  go mod vendor
  cd tests/integration/docker
  mkdir -p coverage
  chmod 777 coverage
  docker compose pull
  docker compose build
  docker compose up -d
  cd ../../..
}

stop_integration() {
  echo "stopping integration tests services"

  cd tests/integration/docker
  docker compose down
  cd ../../..
}

start_integration_debug() {
  echo "starting integration debug tests services"

  touch ./tests/integration/integration.lock
  cd tests/integration/docker
  mkdir -p coverage
  chmod 777 coverage
  docker compose up -d
  cd ../../..
}

stop_integration_debug() {
  echo "stopping integration debug tests services"

  cd tests/integration/docker
  docker compose down
  cd ../../..
  rm ./tests/integration/integration.lock
}

restart_integration_debug() {
  echo "restarting integration debug tests services"

  start_integration
  touch ./tests/integration/integration.lock
  cd tests/integration/docker
  mkdir -p coverage
  chmod 777 coverage
  docker compose up -d
  cd ../../..
}

# Запуск интеграционных тестов
integration(){
  echo "run integration tests"
  start_integration
  touch ./tests/integration/integration.lock
  set +e
  go test ./tests/integration/... -v
  local exit=$?
  set -e
  rm ./tests/integration/integration.lock
  stop_integration

  if [[ $exit != 0 ]]
  then
    echo -e "${RED}integration tests failed${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}integration tests passed${NC}"
  fi
}

integration_docker(){
  echo "run integration tests"
  docker build --file=Dockerfile.test -t $REPO_NAME-test:local .
  start_integration
  touch ./tests/integration/integration.lock
  go mod vendor
  set +e
  docker run --rm -v "$(pwd)":/work -w /work \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --network=host \
    -it $REPO_NAME-test:local \
    bash -c "go test ./tests/integration/... -v"
  local exit=$?
  set -e
  rm ./tests/integration/integration.lock
  rm -Rf vendor
  stop_integration

  if [[ $exit != 0 ]]
  then
    echo -e "${RED}integration tests failed${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}integration tests passed${NC}"
  fi
}

# Запуск тестов dev сервера
test_dev(){
  echo "run dev server tests"
  touch ./tests/server/dev.lock
  set +e
  go test ./tests/server/... -v
  local exit=$?
  set -e
  rm ./tests/server/dev.lock

  if [[ $exit != 0 ]]
  then
    echo -e "${RED}dev server tests failed${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}dev server tests passed${NC}"
    print_success
  fi
}

# Запуск тестов dev сервера
test_stage(){
  echo "run stage server tests"
  touch ./tests/server/stage.lock
  set +e
  go test ./tests/server/... -v
  local exit=$?
  set -e
  rm ./tests/server/stage.lock

  if [[ $exit != 0 ]]
  then
    echo -e "${RED}stage server tests failed${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}stage server tests passed${NC}"
    print_success
  fi
}

# Запуск тестов prod сервера
test_prod(){
  echo "run prod server tests"
  touch ./tests/server/prod.lock
  set +e
  go test ./tests/server/... -v
  local exit=$?
  set -e
  rm ./tests/server/prod.lock

  if [[ $exit != 0 ]]
  then
    echo -e "${RED}prod server tests failed${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}prod server tests failed${NC}"
    print_success
  fi
}

# Подтянуть зависимости
deps(){
  go mod download
}

# Собрать исполняемый файл
build(){
  deps
  CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o app ./app
}

# Запустить сбор метрик нагрузки на cpu из pprof
pprof_cpu(){
  local SECS=${3:-$PPROF_DEFAULT_CPU_DURATION}
  local HOST=$2

  go tool pprof -http :$PPROF_UI_PORT $HOST/debug/pprof/profile?seconds=$SECS
}

# Запустить сбор метрик памяти из pprof
pprof_heap(){
  local HOST=$2

  go tool pprof -http :$PPROF_UI_PORT $HOST/debug/pprof/heap
}

# Собрать docker образ
build_docker() {
  build
  docker build -t "$REPO_NAME:local" .
  rm ./app/app
}

# Запустить проверку локального образа на уязвимости
security_scan() {
  echo "run security scan"
  build_docker
  docker save "$REPO_NAME:local" > image.tar
  docker run --rm -it -v "$(pwd):/work" aquasec/trivy image --input /work/image.tar --timeout 10m0s
  rm image.tar
}

# Запуск генерации конфигов деплоя
gen_config(){
  cd tools/gen_config
  go run .
}

unit_cover() {
  rm -rf coverage/unit
  mkdir -p coverage/unit
  chmod 777 coverage/unit
  # run unit tests with cover option
  if ! go test -cover -coverpkg=./... $(go list ./...) -args -test.gocoverdir="$PWD/coverage/unit";
  then
    echo -e "${RED}[UNIT TESTS FAILED]${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}[UNIT TESTS PASSED]${NC}"
  fi
}

generate_coverage_files() {
  # generate coverage for integration tests
  go tool covdata textfmt -i=./tests/integration/docker/coverage -o ./coverage/cover_profile.out.tmp
  # remove generated code and mocks from coverage
  < ./coverage/cover_profile.out.tmp grep -v -e "mock" > ./coverage/cover_profile.out.tmp2
  < ./coverage/cover_profile.out.tmp2 grep -v -e "authentication/tests" > ./coverage/cover_profile.out
  # generate integration-tests html coverage file
  go tool cover -html=./coverage/cover_profile.out -o ./coverage/cover_integration.html
  echo INTEGRATION_COVERAGE=$( go tool cover -func=./coverage/cover_profile.out | tail -n 1 | awk '{ print $3 }' | sed -e 's/^\([0-9]*\).*$/\1/g' )

  # delete
  rm ./coverage/cover_profile.out.tmp
  rm ./coverage/cover_profile.out.tmp2
  rm ./coverage/cover_profile.out

  # generate coverage for integration unit
  go tool covdata textfmt -i=./coverage/unit -o ./coverage/cover_profile.out.tmp
  # remove generated code and mocks from coverage
  < ./coverage/cover_profile.out.tmp grep -v -e "mock" > ./coverage/cover_profile.out.tmp2
  < ./coverage/cover_profile.out.tmp2 grep -v -e "authentication/tests" > ./coverage/cover_profile.out
  # generate unit-tests html coverage file
  go tool cover -html=./coverage/cover_profile.out -o ./coverage/cover_unit.html
  echo UNIT_COVERAGE=$( go tool cover -func=./coverage/cover_profile.out | tail -n 1 | awk '{ print $3 }' | sed -e 's/^\([0-9]*\).*$/\1/g' )

  rm ./coverage/cover_profile.out.tmp
  rm ./coverage/cover_profile.out.tmp2
  rm ./coverage/cover_profile.out

  # generate total coverage
  go tool covdata textfmt -i=./coverage/unit/,./tests/integration/docker/coverage/ -o coverage/cover_profile.out.tmp
  # remove generated code and mocks from coverage
  < ./coverage/cover_profile.out.tmp grep -v -e "mock" > ./coverage/cover_profile.out.tmp2
  < ./coverage/cover_profile.out.tmp2 grep -v -e "authentication/tests" > ./coverage/cover_profile.out
  # generate html total coverage file
  go tool cover -html=./coverage/cover_profile.out -o ./coverage/cover_total.html

  rm ./coverage/cover_profile.out.tmp
  rm ./coverage/cover_profile.out.tmp2

  # delete all cov-files collected during unit and integration tests
  rm ./coverage/unit/covmeta.*
  rm ./coverage/unit/covcounters.*
  rm -f ./tests/integration/docker/coverage/covmeta.*
  rm -f ./tests/integration/docker/coverage/covcounters.*
}

count_coverage() {
  unit_cover
  generate_coverage_files
  # grep current coverage from total cover profile
  CUR_COVERAGE=$( go tool cover -func=./coverage/cover_profile.out | tail -n 1 | awk '{ print $3 }' | sed -e 's/^\([0-9]*\).*$/\1/g' )
  echo TOTAL_COVERAGE=$CUR_COVERAGE
  rm ./coverage/cover_profile.out
  if [ "$CUR_COVERAGE" -lt $COVERAGE_MIN ]
    then
      echo -e "${RED}coverage is not enough $CUR_COVERAGE < $COVERAGE_MIN ${NC}"
      print_fail
      return 1
    else
      echo -e "${GREEN}coverage is enough $CUR_COVERAGE >= $COVERAGE_MIN${NC}"
    fi
  echo "Для получения детальной информации смотрите /coverage/*.html"
}

# https://patorjk.com/software/taag/#p=display&f=ANSI%20Regular
print_fail() {
  echo -e "${RED} ███████  █████   ████  ██     ${NC}"
  echo -e "${RED} ██      ██   ██   ██   ██     ${NC}"
  echo -e "${RED} █████   ███████   ██   ██     ${NC}"
  echo -e "${RED} ██      ██   ██   ██   ██     ${NC}"
  echo -e "${RED} ██      ██   ██  ████  ██████ ${NC}"
}

print_success() {
  echo -e "${GREEN} ███████  ██    ██   ██████   ██████  ███████  ███████  ███████ ${NC}"
  echo -e "${GREEN} ██       ██    ██  ██       ██       ██       ██       ██      ${NC}"
  echo -e "${GREEN} ███████  ██    ██  ██       ██       █████    ███████  ███████ ${NC}"
  echo -e "${GREEN}      ██  ██    ██  ██       ██       ██            ██       ██ ${NC}"
  echo -e "${GREEN} ███████   ██████    ██████   ██████  ███████  ███████  ███████ ${NC}"
}

githooks() {
  rm -f .git/hooks/pre-commit
  cp .githooks/pre-commit .git/hooks
}

# Запуск всех тестов
test(){
  fmt
  vet
  unit
  deps_check
  lint
  security_scan
  integration
  count_coverage
  print_success
}

test_docker() {
  fmt
  vet
  unit_docker
  deps_check
  lint
  security_scan
  integration_docker
}

# Добавьте сюда список команд
using(){
  echo "Укажите команду при запуске: ./run.sh [command]"
  echo "Список команд:"
  echo "  unit - запустить unit-тесты с проверкой на data-race"
  echo "  unit_docker - запуск unit тестов и проверка покрытия кода тестами в докере (для ос, отличных от линукса)"
  echo "  unit_cover - запуск unit тестов с генерацией cov-файлов для подсчёта покрытия"
  echo "  integration - запуск интеграционных тестов"
  echo "  integration_docker - запуск интеграционных тестов в докере (для ос, отличных от линукса)"
  echo "  lint - запустить все линтеры"
  echo "  test - запустить все тесты"
  echo "  test_docker - запустить все тесты в докере (для ос, отличных от линукса)"
  echo "  count_coverage - запустить подсчёт покрытия (перед этим нужно запустить интеграционные тесты)"
  echo "  deps - подтянуть зависимости"
  echo "  build - собрать приложение"
  echo "  build_docker - собрать локальный docker образ"
  echo "  fmt - форматирование кода при помощи 'go fmt'"
  echo "  vet - проверка правильности форматирования кода"
  echo "  deps_check - анализ зависимостей на уязвимости"
  echo "  security_scan - запустить проверку локального образа на уязвимости"
  echo "  pprof_cpu HOST [SECONDS] - сбор метрик нагрузки на cpu из pprof"
  echo "  pprof_heap HOST - запустить сбор метрик памяти из pprof"
  echo "  start_local - запустить необходимые для локальной разработки сервисы"
  echo "  stop_local - остановить необходимые для локальной разработки сервисы"
  echo "  start_integration - запустить необходимые для интеграционных тестов сервисы"
  echo "  stop_integration - остановить необходимые для интеграционных тестов сервисы"
  echo "  start_integration_debug - остановить необходимые для интеграционных тестов сервисы в режиме debug"
  echo "  stop_integration_debug - остановить необходимые для интеграционных тестов сервисы  в режиме debug"
  echo "  test_dev - запуск тестов dev сервера"
  echo "  test_stage - запуск тестов stage сервера"
  echo "  test_prod - запуск тестов prod сервера"
  echo "  gen_config - запуск генерации конфигов деплоя"
  echo "  githooks - включить git hooks"
}

############### НЕ МЕНЯЙТЕ КОД НИЖЕ ЭТОЙ СТРОКИ #################

command="$1"
if [ -z "$command" ]
then
 using
 exit 0;
else
 $command $@
fi