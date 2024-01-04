#!/bin/bash -e

REPO_NAME=todos
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

lint(){
  echo "run linter"
  if ! docker run --rm -v "$(pwd)":/work:ro -w /work -it golangci/golangci-lint:latest golangci-lint run -v
  then
    echo -e "${RED}[LINTER CHECK FAILED]${NC}"
    print_fail
    return 1
  else
    echo -e "${GREEN}[LINTER CHECK PASSED]${NC}"
  fi
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

# Подтянуть зависимости
deps(){
  go mod download
}

# Запустить проверку локального образа на уязвимости
security_scan() {
  echo "run security scan"
  build_docker
  docker save "$REPO_NAME:local" > image.tar
  docker run --rm -it -v "$(pwd):/work" aquasec/trivy image --input /work/image.tar --timeout 10m0s
  rm image.tar
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