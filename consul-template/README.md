## Demo Tags

* Config Binding - cfgBinding
* Service Binding - svcBinding

## Consul Setup

* Native Mac OS X docker - pull images while disconnect from the FMN
* Need to specify a client-addr to expose the UI outside the container

<pre>
docker run -p 8500:8500 consul agent -dev -client=0.0.0.0
</pre>

* Dev cluster

<pre>
docker run -d --name=dev-consul -p 8500:8500 consul agent -dev -client=0.0.0.0
docker run -d consul agent -dev -join=172.17.0.2
docker run -d consul agent -dev -join=172.17.0.2
</pre>

* Cluster members

<pre>
docker exec -t dev-consul consul members
</pre>

## Mountebank Setup

<pre>
curl -i -X POST -d@endpoint-setup.json http://127.0.0.1:2525/imposters
curl -i -X POST -d@endpoint2-setup.json http://127.0.0.1:2525/imposters
</pre>

## Run with environment variables - cfgBinding

<pre>
docker run -e "endpoint=foo:4545" -e "port=3000" --link mountebank:foo -p 3000:3000  1ac129181e49
</pre>

## Run with environment variables - svcBinding

<pre>
docker run -e "port=3000" -e "env=env1" --link mountebank:mbhost --link dev-consul:consul -p 3000:3000 svcsample
</pre>

## Golang Setup for Demo

* golang set up - consul.sh

## Consul Env

* Install go get github.com/mitchellh/gox, clone https://github.com/hashicorp/envconsul,
and make bin

<pre>
./envconsul  -consul=localhost:8500 -once -prefix=sample -pristine -upcase env
</pre>


## Consul Template

* Project - https://github.com/hashicorp/consul-template
* Releases - https://releases.hashicorp.com/consul-template/

## Docker Image

Build image for tag cfgBinding thusly:

<pre>
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o service
docker build -t cfgsample .
</pre>

Build image for tag svcBinding thusly:

<pre>
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o service
docker build -t svcsample .
</pre>


## Add config

<pre>
curl -X PUT localhost:8500/v1/kv/env1/
curl -X PUT localhost:8500/v1/kv/env1/endpoint-host -d foo
curl -X PUT localhost:8500/v1/kv/env1/endpoint-port -d 4545
curl -X PUT localhost:8500/v1/kv/env1/port -d 3000
</pre>

## Consul-template

<pre>
./consul-template -consul localhost:8500 -template $HOME/goex/src/github.com/d-smith/go-examples/consul-template/demo-template.ctmpl -dry -once
</pre>

## Run mountebank

<pre>
docker pull dasmith/mb-server-alpine
docker run -d -p 2525:2525 --name mountebank dasmith/mb-server-alpine
</pre>

## Service Definition

curl -v -X PUT -d@env1service.json localhost:8500/v1/agent/service/register
curl -v -X PUT -d@env2service.json localhost:8500/v1/agent/service/register


## Demo Script

###Env setup

<pre>
. gosetup.sh
cd gostuff/src/github.com/d-smith/go-examples/consul-template/
export no_proxy=127.0.0.1,localhost
</pre>

###Clean up

<pre>
docker rm $(docker ps -aq)
</pre>


### Mountebank

Start mountebank, and create the mock service endpoints:

<pre>
docker run -d -p 2525:2525 -p 4545:4545 -p 5555:5555 --name mountebank dasmith/mb-server-alpine
curl -i -X POST -d@endpoint-setup.json http://127.0.0.1:2525/imposters
curl -i -X POST -d@endpoint2-setup.json http://127.0.0.1:2525/imposters

curl localhost:4545
curl localhost:5555
</pre>

### Consul Template

<pre>
docker run -d --name=dev-consul -p 8500:8500 consul agent -dev -client=0.0.0.0
</pre>

Show UI:

<pre>
http://localhost:8500/ui/
</pre>

Add configuration:

<pre>
curl -X PUT localhost:8500/v1/kv/env1/
curl -X PUT localhost:8500/v1/kv/env1/endpoint-host -d foo
curl -X PUT localhost:8500/v1/kv/env1/endpoint-port -d 4545
curl -X PUT localhost:8500/v1/kv/env1/port -d 3000
</pre>

Generate run script from config store:

<pre>
./consul-template -consul localhost:8500 -template demo-template.ctmpl -dry -once

cat demo-template.ctmpl
</pre>

### Service Discovery

Show service.go

Load service definitions:

<pre>
curl -v -X PUT -d@env1service.json localhost:8500/v1/agent/service/register
curl -v -X PUT -d@env2service.json localhost:8500/v1/agent/service/register
</pre>

Set up health checks:

<pre>
curl -v -X PUT -d@env1hc.json localhost:8500/v1/agent/check/register
</pre>

Run service for env 1:

<pre>
docker run -e "port=3000" -e "env=env1" --link mountebank:mbhost --link dev-consul:consul -p 3000:3000 svcsample
docker run -e "port=3000" -e "env=env2" --link mountebank:mbhost --link dev-consul:consul -p 3000:3000 svcsample
</pre>
